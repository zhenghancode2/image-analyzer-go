package imageutil

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"image-analyzer-go/pkg/utils"

	"github.com/containers/image/v5/copy"
	"github.com/containers/image/v5/docker"
	"github.com/containers/image/v5/oci/layout"
	"github.com/containers/image/v5/signature"
	"github.com/containers/image/v5/types"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

// generateUniqueID 生成一个唯一的标识符
func generateUniqueID() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("生成随机数失败: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// PullAndExtract 从指定的镜像引用中提取镜像层
// 返回提取的目录路径和镜像配置，如果发生错误则返回错误
func PullAndExtract(ctx context.Context, refStr string) (string, *v1.Image, error) {
	sys := &types.SystemContext{}
	ref, err := docker.ParseReference("//" + refStr)
	if err != nil {
		return "", nil, utils.WrapError(err, "解析镜像引用失败")
	}

	tmpDir, err := utils.CreateTempDir("oci-layout")
	if err != nil {
		return "", nil, err
	}
	defer func() {
		if cleanupErr := utils.CleanupTempDir(tmpDir); cleanupErr != nil {
			fmt.Printf("警告: 清理临时目录 %s 失败: %v\n", tmpDir, cleanupErr)
		}
	}()

	dest, err := layout.NewReference(tmpDir, "latest")
	if err != nil {
		return "", nil, utils.WrapError(err, "创建布局引用失败")
	}

	policy, err := signature.NewPolicyContext(&signature.Policy{Default: []signature.PolicyRequirement{signature.NewPRInsecureAcceptAnything()}})
	if err != nil {
		return "", nil, utils.WrapError(err, "创建策略上下文失败")
	}
	defer policy.Destroy()

	_, err = copy.Image(ctx, policy, ref, dest, &copy.Options{})
	if err != nil {
		return "", nil, utils.WrapError(err, "复制镜像失败")
	}

	img, err := dest.NewImage(ctx, sys)
	if err != nil {
		return "", nil, utils.WrapError(err, "创建镜像失败")
	}
	defer img.Close()

	cfg, err := img.OCIConfig(ctx)
	if err != nil {
		return "", nil, utils.WrapError(err, "获取OCI配置失败")
	}

	extractDir, err := utils.CreateTempDir("layers")
	if err != nil {
		return "", nil, err
	}

	layers := img.LayerInfos()
	for _, layer := range layers {
		blobPath := filepath.Join(tmpDir, "blobs", "sha256", layer.Digest.Hex())
		f, err := os.Open(blobPath)
		if err != nil {
			if cleanupErr := utils.CleanupTempDir(extractDir); cleanupErr != nil {
				fmt.Printf("警告: 清理提取目录 %s 失败: %v\n", extractDir, cleanupErr)
			}
			return "", nil, utils.WrapError(err, "打开层文件失败")
		}
		decompressAndUntar(f, extractDir)
		f.Close()
	}

	return extractDir, cfg, nil
}

func decompressAndUntar(r io.Reader, dest string) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return
	}
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil || hdr == nil {
			break
		}
		path := filepath.Join(dest, hdr.Name)
		if hdr.Typeflag == tar.TypeDir {
			os.MkdirAll(path, 0755)
		} else if hdr.Typeflag == tar.TypeReg {
			os.MkdirAll(filepath.Dir(path), 0755)
			f, _ := os.Create(path)
			io.Copy(f, tr)
			f.Close()
		}
	}
}
