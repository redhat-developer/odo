package occlient

import (
	"context"
	"fmt"
	"io"
	"time"

	buildv1 "github.com/openshift/api/build/v1"
	buildschema "github.com/openshift/client-go/build/clientset/versioned/scheme"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/util"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/klog"
)

// GetBuildConfigFromName get BuildConfig by its name
func (c *Client) GetBuildConfigFromName(name string) (*buildv1.BuildConfig, error) {
	klog.V(3).Infof("Getting BuildConfig: %s", name)
	bc, err := c.buildClient.BuildConfigs(c.Namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get BuildConfig %s", name)
	}
	return bc, nil
}

// GetLatestBuildName gets the name of the latest build
// buildConfigName is the name of the buildConfig for which we are fetching the build name
// returns the name of the latest build or the error
func (c *Client) GetLatestBuildName(buildConfigName string) (string, error) {
	buildConfig, err := c.buildClient.BuildConfigs(c.Namespace).Get(context.TODO(), buildConfigName, metav1.GetOptions{})
	if err != nil {
		return "", errors.Wrap(err, "unable to get the latest build name")
	}
	return fmt.Sprintf("%s-%d", buildConfigName, buildConfig.Status.LastVersion), nil
}

// CreateBuildConfig creates a buildConfig using the builderImage as well as gitURL.
// envVars is the array containing the environment variables
func (c *Client) CreateBuildConfig(commonObjectMeta metav1.ObjectMeta, builderImage string, gitURL string, gitRef string, envVars []corev1.EnvVar) (buildv1.BuildConfig, error) {

	// Retrieve the namespace, image name and the appropriate tag
	imageNS, imageName, imageTag, _, err := ParseImageName(builderImage)
	if err != nil {
		return buildv1.BuildConfig{}, errors.Wrap(err, "unable to parse image name")
	}
	imageStream, err := c.GetImageStream(imageNS, imageName, imageTag)
	if err != nil {
		return buildv1.BuildConfig{}, errors.Wrap(err, "unable to retrieve image stream for CreateBuildConfig")
	}
	imageNS = imageStream.ObjectMeta.Namespace

	klog.V(3).Infof("Using namespace: %s for the CreateBuildConfig function", imageNS)

	// Use BuildConfig to build the container with Git
	bc := generateBuildConfig(commonObjectMeta, gitURL, gitRef, imageName+":"+imageTag, imageNS)

	if len(envVars) > 0 {
		bc.Spec.Strategy.SourceStrategy.Env = envVars
	}
	_, err = c.buildClient.BuildConfigs(c.Namespace).Create(context.TODO(), &bc, metav1.CreateOptions{FieldManager: kclient.FieldManager})
	if err != nil {
		return buildv1.BuildConfig{}, errors.Wrapf(err, "unable to create BuildConfig for %s", commonObjectMeta.Name)
	}

	return bc, nil
}

// UpdateBuildConfig updates the BuildConfig file
// buildConfigName is the name of the BuildConfig file to be updated
// projectName is the name of the project
// gitURL equals to the git URL of the source and is equals to "" if the source is of type dir or binary
// annotations contains the annotations for the BuildConfig file
func (c *Client) UpdateBuildConfig(buildConfigName string, gitURL string, annotations map[string]string) error {
	if gitURL == "" {
		return errors.New("gitURL for UpdateBuildConfig must not be blank")
	}

	// generate BuildConfig
	buildSource := buildv1.BuildSource{
		Git: &buildv1.GitBuildSource{
			URI: gitURL,
		},
		Type: buildv1.BuildSourceGit,
	}

	buildConfig, err := c.GetBuildConfigFromName(buildConfigName)
	if err != nil {
		return errors.Wrap(err, "unable to get the BuildConfig file")
	}
	buildConfig.Spec.Source = buildSource
	buildConfig.Annotations = annotations
	_, err = c.buildClient.BuildConfigs(c.Namespace).Update(context.TODO(), buildConfig, metav1.UpdateOptions{FieldManager: kclient.FieldManager})
	if err != nil {
		return errors.Wrap(err, "unable to update the component")
	}
	return nil
}

// DeleteBuildConfig deletes the given BuildConfig by name using CommonObjectMeta..
func (c *Client) DeleteBuildConfig(commonObjectMeta metav1.ObjectMeta) error {

	// Convert labels to selector
	selector := util.ConvertLabelsToSelector(commonObjectMeta.Labels)
	klog.V(3).Infof("DeleteBuildConfig selectors used for deletion: %s", selector)

	// Delete BuildConfig
	klog.V(3).Info("Deleting BuildConfigs with DeleteBuildConfig")
	return c.buildClient.BuildConfigs(c.Namespace).DeleteCollection(context.TODO(), metav1.DeleteOptions{}, metav1.ListOptions{LabelSelector: selector})
}

// WaitForBuildToFinish block and waits for build to finish. Returns error if build failed or was canceled.
func (c *Client) WaitForBuildToFinish(buildName string, stdout io.Writer, buildTimeout time.Duration) error {
	// following indicates if we have already setup the following logic
	following := false
	klog.V(3).Infof("Waiting for %s  build to finish", buildName)

	// start a watch on the build resources and look for the given build name
	w, err := c.buildClient.Builds(c.Namespace).Watch(context.TODO(), metav1.ListOptions{
		FieldSelector: fields.Set{"metadata.name": buildName}.AsSelector().String(),
	})
	if err != nil {
		return errors.Wrapf(err, "unable to watch build")
	}
	defer w.Stop()
	timeout := time.After(buildTimeout)
	for {
		select {
		// when a event is received regarding the given buildName
		case val, ok := <-w.ResultChan():
			if !ok {
				break
			}
			// cast the object returned to a build object and check the phase of the build
			if e, ok := val.Object.(*buildv1.Build); ok {
				klog.V(3).Infof("Status of %s build is %s", e.Name, e.Status.Phase)
				switch e.Status.Phase {
				case buildv1.BuildPhaseComplete:
					// the build is completed thus return
					klog.V(3).Infof("Build %s completed.", e.Name)
					return nil
				case buildv1.BuildPhaseFailed, buildv1.BuildPhaseCancelled, buildv1.BuildPhaseError:
					// the build failed/got cancelled/error occurred thus return with error
					return errors.Errorf("build %s status %s", e.Name, e.Status.Phase)
				case buildv1.BuildPhaseRunning:
					// since the pod is ready and the build is now running, start following the logs
					if !following {
						// setting following to true as we need to set it up only once
						following = true
						err := c.FollowBuildLog(buildName, stdout, buildTimeout)
						if err != nil {
							return err
						}
					}
				}
			}
		case <-timeout:
			// timeout has occurred while waiting for the build to start/complete, so error out
			return errors.Errorf("timeout waiting for build %s to start", buildName)
		}
	}
}

// StartBuild starts new build as it is, returns name of the build stat was started
func (c *Client) StartBuild(name string) (string, error) {
	klog.V(3).Infof("Build %s started.", name)
	buildRequest := buildv1.BuildRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	result, err := c.buildClient.BuildConfigs(c.Namespace).Instantiate(context.TODO(), name, &buildRequest, metav1.CreateOptions{FieldManager: kclient.FieldManager})
	if err != nil {
		return "", errors.Wrapf(err, "unable to instantiate BuildConfig for %s", name)
	}
	klog.V(3).Infof("Build %s for BuildConfig %s triggered.", name, result.Name)

	return result.Name, nil
}

// FollowBuildLog stream build log to stdout
func (c *Client) FollowBuildLog(buildName string, stdout io.Writer, buildTimeout time.Duration) error {
	buildLogOptions := buildv1.BuildLogOptions{
		Follow: true,
		NoWait: false,
	}

	rd, err := c.buildClient.RESTClient().Get().
		Timeout(buildTimeout).
		Namespace(c.Namespace).
		Resource("builds").
		Name(buildName).
		SubResource("log").
		VersionedParams(&buildLogOptions, buildschema.ParameterCodec).
		Stream(context.TODO())

	if err != nil {
		return errors.Wrapf(err, "unable get build log %s", buildName)
	}
	defer rd.Close()

	if _, err = io.Copy(stdout, rd); err != nil {
		return errors.Wrapf(err, "error streaming logs for %s", buildName)
	}

	return nil
}
