package domain_test

import (
	"encoder/domain"
	"testing"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
)

func TestValidateIfVideoIsEmpty(t *testing.T) {
	video := domain.NewVideo()
	err := video.Validade()

	require.Error(t, err)
}

func TestVideoIdIsNotAUuid(t *testing.T) {
	video := domain.Video{}

	video.ID = "abc"
	video.ResourceID = "a"
	video.FilePath = "path"
	video.CreatedAt = time.Now()

	err := video.Validade()

	require.Error(t, err)

}

func TestVideoValidation(t *testing.T) {
	video := domain.Video{}

	video.ID = uuid.NewV4().String()
	video.ResourceID = "a"
	video.FilePath = "path"
	video.CreatedAt = time.Now()

	err := video.Validade()

	require.Nil(t, err)

}
