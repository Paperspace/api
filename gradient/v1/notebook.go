package v1

import (
	"encoding/base64"
	"strings"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate controller-gen object paths=$GOFILE

// NotebookSpec defines the desired state of Notebook
// +k8s:openapi-gen=true
type NotebookSpec struct {
	Name               string            `json:"name,required"`
	Workspace          *Workspace        `json:"workspace,omitempty"`
	ProjectHandle      string            `json:"projectHandle,required"`
	TeamHandle         string            `json:"teamHandle,required"`
	UserHandle         string            `json:"userHandle,required"`
	Handle             string            `json:"handle,required"`
	JobHandle          string            `json:"jobHandle,required"`
	NotebookRepoHandle string            `json:"notebookRepoHandle,omitempty"`
	Token              string            `json:"token,required"`
	APIKey             string            `json:"apiKey,omitempty"`
	TTL                int               `json:"TTL,omitempty"`
	Upload             NotebookUpload    `json:"upload,omitempty"`
	Instance           Instance          `json:"instance,required"`
	Details            NotebookDetails   `json:"details,required"`
	Env                map[string]string `json:"env,omitempty"`
	VolumeMounts       []VolumeMount     `json:"volumeMounts,omitempty"`
}

type NotebookUpload struct {
	S3Upload    S3Upload     `json:"s3Upload,omitempty"`
	ImageUpload *ImageUpload `json:"imageUpload,omitempty"`
}

type ImageUpload struct {
	Registry   string `json:"registry,omitempty"`
	Repository string `json:"repository,omitempty"`
	Username   string `json:"username,omitempty"`
	Password   string `json:"password,omitempty"`
}

func (iu *ImageUpload) HasCredentials() bool {
	return iu != nil
}

type DeserializedImageUpload ImageUpload

func (iu *ImageUpload) GetDeserializedCredentials() DeserializedImageUpload {
	if !iu.HasCredentials() {
		return DeserializedImageUpload{}
	}
	username, _ := base64.StdEncoding.DecodeString(iu.Username)
	password, _ := base64.StdEncoding.DecodeString(iu.Password)
	registry, _ := base64.StdEncoding.DecodeString(iu.Registry)
	repository, _ := base64.StdEncoding.DecodeString(iu.Repository)

	return DeserializedImageUpload{
		Registry:   strings.TrimSpace(string(registry)),
		Repository: strings.TrimSpace(string(repository)),
		Username:   strings.TrimSpace(string(username)),
		Password:   strings.TrimSpace(string(password)),
	}
}

type NotebookDetails struct {
	Image   ImageDetails `json:"image,required"`
	Command string       `json:"command,required"`
	WorkDir string       `json:"workDir,omitempty"`
}

type NotebookState string

const (
	_                                  NotebookState = ""
	NotebookStateError                 NotebookState = "Error"
	NotebookStateWaitingForVolume      NotebookState = "WaitingForVolume"
	NotebookStateWaitingForArtifact    NotebookState = "WaitingForArtifact"
	NotebookStateDownloadArtifactError NotebookState = "DownloadArtifactError"
	NotebookStateIngressError          NotebookState = "IngressCreateError"
	NotebookStateServiceError          NotebookState = "ServiceCreateError"
	NotebookStateFinished              NotebookState = "Finished"
	NotebookStatePodStarting           NotebookState = "PodStarting"
	NotebookStateRunning               NotebookState = "Running"
	NotebookStateTearingDown           NotebookState = "Teardown"
)

// NotebookStatus defines the observed state of Notebook
// +k8s:openapi-gen=true
type NotebookStatus struct {
	State                    NotebookState                     `json:"state"`
	GradientStatus           JobGradientStatus                 `json:"gradientStatus"`
	EndpointURL              string                            `json:"endpointURL"`
	Message                  string                            `json:"message,omitempty"`
	ImageSecretName          string                            `json:"imageSecretName"`
	LastUpdatedAt            *metav1.Time                      `json:"lastUpdatedAt,omitempty"`
	RunningAt                *metav1.Time                      `json:"runningAt,omitempty"`
	ExitCode                 int32                             `json:"exitCode"`
	PodStatus                *PodStatus                        `json:"podStatus,omitempty"`
	DownloadArtifactStatuses map[string]DownloadArtifactStatus `json:"downloadArtifactStatuses,omitempty"`
	// NBConvertJobStatus deprecated, retained for backwards compatibility
	NBConvertJobStatus *batchv1.JobStatus `json:"nbconvertJobStatus,omitempty"`
	// WorkspaceUploadJobStatus deprecated, retained for backwards compatibility
	WorkspaceUploadJobStatus *batchv1.JobStatus `json:"workspaceUploadJobStatus,omitempty"`
	// ImageCacheStatus deprecated, retained for backwards compatibility
	ImageCacheJobStatus      *batchv1.JobStatus `json:"imageCacheJobStatus,omitempty"`
	WorkspaceUploadPodStatus *PodStatus         `json:"workspaceUploadPodStatus,omitempty"`
	ImageCachePodStatus      *PodStatus         `json:"imageCachePodStatus,omitempty"`
	ServiceName              string             `json:"serviceName,omitempty"`
	IngressName              string             `json:"ingressName,omitempty"`
	NotebookNodeName         string             `json:"notebookNodeName,omitempty"`

	// WorkspaceExportJobStatus deprecated, retained for backwards compatibility no longer set on notebooks
	WorkspaceExportJobStatus *batchv1.JobStatus `json:"workspaceExportJobStatus,omitempty"`
	// ImageExportJobStatus deprecated, retained for backwards compatibility no longer set on notebooks
	ImageExportJobStatus *batchv1.JobStatus `json:"imageExportJobStatus,omitempty"`
}

var _ Status = (*NotebookStatus)(nil)

func (ns *NotebookStatus) GetLastUpdatedAt() *metav1.Time {
	return ns.LastUpdatedAt
}

func (ns *NotebookStatus) SetLastUpdatedAt(tim metav1.Time) {
	ns.LastUpdatedAt = &tim
}

func (ns *NotebookStatus) IsSuccess() bool {
	return ns.State == NotebookStateFinished
}

func (ns *NotebookStatus) IsErrored() bool {
	return ns.State == NotebookStateError ||
		ns.State == NotebookStateDownloadArtifactError ||
		ns.State == NotebookStateIngressError ||
		ns.State == NotebookStateServiceError
}

func (ns *NotebookStatus) CopyToStatus() Status {
	return ns.DeepCopy()
}

func (ns *NotebookStatus) NeedsGarbageCollection() bool {
	return !ns.PodStatus.IsDeleted() || !ns.WorkspaceUploadPodStatus.IsDeleted() || !ns.ImageCachePodStatus.IsDeleted()
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Notebook is the Schema for the notebooks API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.state`
// +kubebuilder:printcolumn:name="LastUpdatedAt",type="date",JSONPath=`.status.lastUpdatedAt`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=`.metadata.creationTimestamp`
// +kubebuilder:printcolumn:name="RepoHandle",type="string",JSONPath=`.spec.notebookRepoHandle`
type Notebook struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NotebookSpec   `json:"spec,omitempty"`
	Status NotebookStatus `json:"status,omitempty"`
}

func (n *Notebook) SetDefaults() {
	n.Status.GradientStatus.SetDefaults(n.Spec.Handle)
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NotebookList contains a list of Notebook
type NotebookList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Notebook `json:"items"`
}
