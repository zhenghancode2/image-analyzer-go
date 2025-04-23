package imageutil

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"image-analyzer-go/pkg/logger"
	"image-analyzer-go/pkg/utils"

	"github.com/containers/image/v5/copy"
	"github.com/containers/image/v5/docker"
	"github.com/containers/image/v5/oci/layout"
	"github.com/containers/image/v5/signature"
	"github.com/containers/image/v5/types"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

// PullAndExtract 从指定的镜像引用中提取镜像层
// 返回提取的目录路径和镜像配置，如果发生错误则返回错误
func PullAndExtract(ctx context.Context, refStr string) (string, *v1.Image, error) {
	sys := &types.SystemContext{
		// 添加 Docker Hub 认证
		DockerAuthConfig: &types.DockerAuthConfig{
			Username: os.Getenv("DOCKER_USERNAME"),
			Password: os.Getenv("DOCKER_PASSWORD"),
		},
		// 跳过 TLS 验证
		DockerInsecureSkipTLSVerify: types.OptionalBoolTrue,
	}

	// 创建临时目录用于存储镜像
	tmpDir, err := utils.CreateTempDir("image-layers")
	if err != nil {
		return "", nil, err
	}

	// 最后清理临时目录
	defer func() {
		if err != nil {
			if cleanupErr := utils.CleanupTempDir(tmpDir); cleanupErr != nil {
				fmt.Printf("警告: 清理临时目录 %s 失败: %v\n", tmpDir, cleanupErr)
			}
		}
	}()

	// 创建源镜像引用
	srcRef, err := docker.ParseReference("//" + refStr)
	if err != nil {
		return "", nil, utils.WrapError(err, "解析源镜像引用失败")
	}

	// 创建目标 OCI 布局引用
	destRef, err := layout.NewReference(tmpDir, "latest")
	if err != nil {
		return "", nil, utils.WrapError(err, "创建目标 OCI 布局引用失败")
	}

	// 创建策略上下文
	policy := &signature.Policy{
		Default: []signature.PolicyRequirement{
			signature.NewPRInsecureAcceptAnything(),
		},
	}
	policyContext, err := signature.NewPolicyContext(policy)
	if err != nil {
		return "", nil, utils.WrapError(err, "创建策略上下文失败")
	}
	defer policyContext.Destroy()

	// 创建进度处理器
	progress := NewProgressHandler()

	// 复制镜像到本地 OCI 布局
	logger.Info("开始下载镜像...")
	_, err = copy.Image(ctx, policyContext, destRef, srcRef, &copy.Options{
		SourceCtx:          sys,
		DestinationCtx:     sys,
		ImageListSelection: copy.CopyAllImages,
		Progress:           progress,
	})
	if err != nil {
		return "", nil, utils.WrapError(err, "复制镜像到本地 OCI 布局失败")
	}
	logger.Info("镜像下载完成")

	// 打开目标镜像
	destImg, err := destRef.NewImage(ctx, sys)
	if err != nil {
		return "", nil, utils.WrapError(err, "打开目标镜像失败")
	}
	defer destImg.Close()

	// 获取镜像配置
	cfg, err := destImg.OCIConfig(ctx)
	if err != nil {
		return "", nil, utils.WrapError(err, "获取OCI配置失败")
	}

	// 获取镜像层
	layers := destImg.LayerInfos()
	logger.Info("开始提取镜像层", logger.WithInt("total_layers", len(layers)))
	for i, layer := range layers {
		logger.Info("开始提取层",
			logger.WithInt("current", i+1),
			logger.WithInt("total", len(layers)),
			logger.WithString("size", utils.FormatBytes(layer.Size)))

		// 创建层目录
		layerDir := filepath.Join(tmpDir, fmt.Sprintf("layer-%d", i))
		if err := os.MkdirAll(layerDir, 0755); err != nil {
			return "", nil, utils.WrapError(err, "创建层目录失败")
		}

		// 获取层内容
		blob, err := destRef.NewImageSource(ctx, sys)
		if err != nil {
			return "", nil, utils.WrapError(err, fmt.Sprintf("创建层 %d 源失败", i))
		}
		defer blob.Close()

		// 读取层数据
		reader, _, err := blob.GetBlob(ctx, layer, nil)
		if err != nil {
			return "", nil, utils.WrapError(err, fmt.Sprintf("获取层 %d 数据失败", i))
		}
		defer reader.Close()

		// 解压并提取层内容
		if err := decompressAndUntar(reader, layerDir); err != nil {
			return "", nil, utils.WrapError(err, fmt.Sprintf("解压层 %d 失败", i))
		}
		logger.Info("层提取完成",
			logger.WithInt("current", i+1),
			logger.WithInt("total", len(layers)))
	}
	logger.Info("所有镜像层提取完成")

	return tmpDir, cfg, nil
}

// decompressAndUntar 解压并提取 tar 文件内容
func decompressAndUntar(r io.Reader, dest string) error {
	// 创建 gzip 读取器
	gz, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("创建 gzip 读取器失败: %w", err)
	}
	defer gz.Close()

	// 创建 tar 读取器
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("读取 tar 头失败: %w", err)
		}

		// 构建目标路径
		path := filepath.Join(dest, hdr.Name)
		if hdr.Typeflag == tar.TypeDir {
			// 创建目录
			if err := os.MkdirAll(path, 0755); err != nil {
				return fmt.Errorf("创建目录失败: %w", err)
			}
		} else if hdr.Typeflag == tar.TypeReg {
			// 创建文件
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				return fmt.Errorf("创建父目录失败: %w", err)
			}
			f, err := os.Create(path)
			if err != nil {
				return fmt.Errorf("创建文件失败: %w", err)
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return fmt.Errorf("复制文件内容失败: %w", err)
			}
			f.Close()
		}
	}
	return nil
}
