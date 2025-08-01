package cdn

import (
	"context"
	"fmt"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type CloudinaryClient struct {
	cld *cloudinary.Cloudinary
	ctx context.Context
}

func NewCloudinaryClient(cloudinaryURL string) (*CloudinaryClient, error) {
	cld, err := cloudinary.NewFromURL(cloudinaryURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create cloudinary client: %w", err)
	}
	
	cld.Config.URL.Secure = true
	
	return &CloudinaryClient{
		cld: cld,
		ctx: context.Background(),
	}, nil
}

func (c *CloudinaryClient) UploadImage(imageURL string, folder string) (string, error) {
	resp, err := c.cld.Upload.Upload(c.ctx, imageURL, uploader.UploadParams{
		Folder:         folder,
		UniqueFilename: api.Bool(true),
		Overwrite:      api.Bool(false),
		ResourceType:   "image",
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload image: %w", err)
	}
	
	return resp.SecureURL, nil
}

func (c *CloudinaryClient) UploadUserAvatar(imageURL string) (string, error) {
	return c.UploadImage(imageURL, "esports_platform/avatars")
}

func (c *CloudinaryClient) UploadKYCDocument(imageURL string) (string, error) {
	return c.UploadImage(imageURL, "esports_platform/kyc")
}

func (c *CloudinaryClient) UploadGameLogo(imageURL string) (string, error) {
	return c.UploadImage(imageURL, "esports_platform/games")
}

func (c *CloudinaryClient) UploadTeamLogo(imageURL string) (string, error) {
	return c.UploadImage(imageURL, "esports_platform/teams")
}

func (c *CloudinaryClient) UploadPlayerAvatar(imageURL string) (string, error) {
	return c.UploadImage(imageURL, "esports_platform/players")
}

func (c *CloudinaryClient) UploadTournamentBanner(imageURL string) (string, error) {
	return c.UploadImage(imageURL, "esports_platform/tournaments")
}

func (c *CloudinaryClient) DeleteImage(publicID string) error {
	_, err := c.cld.Upload.Destroy(c.ctx, uploader.DestroyParams{
		PublicID: publicID,
	})
	return err
}

func (c *CloudinaryClient) GenerateTransformedURL(publicID string, width, height int) string {
	// Simple transformation URL generation
	return fmt.Sprintf("https://res.cloudinary.com/%s/image/upload/w_%d,h_%d/%s", 
		"dwnxysjxp", width, height, publicID)
}