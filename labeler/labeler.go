package labeler

import (
	"fmt"
	"io"

	"golang.org/x/net/context"

	vision "cloud.google.com/go/vision/apiv1"
	gvision "google.golang.org/genproto/googleapis/cloud/vision/v1"
)

var feature = []*gvision.Feature{
	{Type: gvision.Feature_LABEL_DETECTION, MaxResults: 10},
	{Type: gvision.Feature_FACE_DETECTION, MaxResults: 10},
	//	{Type: gvision.Feature_WEB_DETECTION, MaxResults: 10},
}

// Label labels a photo
func Label(file io.Reader) ([]*gvision.EntityAnnotation, []*gvision.FaceAnnotation, *gvision.WebDetection, error) {
	// Use oauth2.NoContext if there isn't a good context to pass in.
	ctx := context.Background()

	// Creates a client.
	client, err := vision.NewImageAnnotatorClient(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Failed to create client: %v", err)
	}

	image, err := vision.NewImageFromReader(file)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Failed to create image: %v", err)
	}

	res, err := client.AnnotateImage(ctx, &gvision.AnnotateImageRequest{
		Image:        image,
		ImageContext: nil,
		Features:     feature,
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Failed to annotate image: %v", err)
	}

	return res.LabelAnnotations, res.FaceAnnotations, res.WebDetection, nil
}
