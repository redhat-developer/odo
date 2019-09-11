package testingutil

import (
	imagev1 "github.com/openshift/api/image/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Function taken from occlient_test.go
// fakeImageStream gets imagestream for the reactor
func fakeImageStream(imageName string, namespace string, tags []string) *imagev1.ImageStream {
	image := &imagev1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name:      imageName,
			Namespace: namespace,
		},
		Status: imagev1.ImageStreamStatus{
			Tags: []imagev1.NamedTagEventList{
				{
					Tag: "latest",
					Items: []imagev1.TagEvent{
						{DockerImageReference: "example/" + imageName + ":latest"},
						{Generation: 1},
						{Image: imageName + "@sha256:9579a93ee"},
					},
				},
			},
		},
	}

	for _, tag := range tags {
		imageTag := imagev1.TagReference{
			Name:        tag,
			Annotations: map[string]string{"tags": "builder"},
		}
		image.Spec.Tags = append(image.Spec.Tags, imageTag)
	}

	return image
}

// FakeImageStreams lists the imagestreams for the reactor
func FakeImageStreams(imageName string, namespace string, tags []string) *imagev1.ImageStreamList {
	return &imagev1.ImageStreamList{
		Items: []imagev1.ImageStream{*fakeImageStream(imageName, namespace, tags)},
	}
}

// fakeImageStreamTag gets imagestreamtag for the reactor
func fakeImageStreamTag(imageName string, version string, namespace string, isHidden bool) *imagev1.ImageStreamTag {
	tagsStr := ""
	if isHidden {
		tagsStr = "hidden"
	}
	image := &imagev1.ImageStreamTag{
		ObjectMeta: metav1.ObjectMeta{
			Name:      imageName + ":" + version,
			Namespace: namespace,
			Annotations: map[string]string{
				"tags":    tagsStr,
				"version": version,
			},
		},
	}

	return image
}

// FakeImageStreamTags lists the imagestreams for the reactor
func FakeImageStreamTags(imageName string, namespace string, tags []string, hiddenTags []string) *imagev1.ImageStreamTagList {
	var list = []imagev1.ImageStreamTag{}
	for _, tag := range tags {
		isHidden := false
		for _, ht := range hiddenTags {
			if ht == tag {
				isHidden = true
				break
			}
		}
		list = append(list, *fakeImageStreamTag(imageName, tag, namespace, isHidden))
	}

	return &imagev1.ImageStreamTagList{
		Items: list,
	}
}
