// Code generated by cdpgen. DO NOT EDIT.

package heapprofiler

import (
	"github.com/mafredri/cdp/protocol/runtime"
)

// StartTrackingHeapObjectsArgs represents the arguments for StartTrackingHeapObjects in the HeapProfiler domain.
type StartTrackingHeapObjectsArgs struct {
	TrackAllocations *bool `json:"trackAllocations,omitempty"` // No description.
}

// NewStartTrackingHeapObjectsArgs initializes StartTrackingHeapObjectsArgs with the required arguments.
func NewStartTrackingHeapObjectsArgs() *StartTrackingHeapObjectsArgs {
	args := new(StartTrackingHeapObjectsArgs)

	return args
}

// SetTrackAllocations sets the TrackAllocations optional argument.
func (a *StartTrackingHeapObjectsArgs) SetTrackAllocations(trackAllocations bool) *StartTrackingHeapObjectsArgs {
	a.TrackAllocations = &trackAllocations
	return a
}

// StopTrackingHeapObjectsArgs represents the arguments for StopTrackingHeapObjects in the HeapProfiler domain.
type StopTrackingHeapObjectsArgs struct {
	ReportProgress *bool `json:"reportProgress,omitempty"` // If true 'reportHeapSnapshotProgress' events will be generated while snapshot is being taken when the tracking is stopped.
}

// NewStopTrackingHeapObjectsArgs initializes StopTrackingHeapObjectsArgs with the required arguments.
func NewStopTrackingHeapObjectsArgs() *StopTrackingHeapObjectsArgs {
	args := new(StopTrackingHeapObjectsArgs)

	return args
}

// SetReportProgress sets the ReportProgress optional argument. If true 'reportHeapSnapshotProgress' events will be generated while snapshot is being taken when the tracking is stopped.
func (a *StopTrackingHeapObjectsArgs) SetReportProgress(reportProgress bool) *StopTrackingHeapObjectsArgs {
	a.ReportProgress = &reportProgress
	return a
}

// TakeHeapSnapshotArgs represents the arguments for TakeHeapSnapshot in the HeapProfiler domain.
type TakeHeapSnapshotArgs struct {
	ReportProgress *bool `json:"reportProgress,omitempty"` // If true 'reportHeapSnapshotProgress' events will be generated while snapshot is being taken.
}

// NewTakeHeapSnapshotArgs initializes TakeHeapSnapshotArgs with the required arguments.
func NewTakeHeapSnapshotArgs() *TakeHeapSnapshotArgs {
	args := new(TakeHeapSnapshotArgs)

	return args
}

// SetReportProgress sets the ReportProgress optional argument. If true 'reportHeapSnapshotProgress' events will be generated while snapshot is being taken.
func (a *TakeHeapSnapshotArgs) SetReportProgress(reportProgress bool) *TakeHeapSnapshotArgs {
	a.ReportProgress = &reportProgress
	return a
}

// GetObjectByHeapObjectIDArgs represents the arguments for GetObjectByHeapObjectID in the HeapProfiler domain.
type GetObjectByHeapObjectIDArgs struct {
	ObjectID    HeapSnapshotObjectID `json:"objectId"`              // No description.
	ObjectGroup *string              `json:"objectGroup,omitempty"` // Symbolic group name that can be used to release multiple objects.
}

// NewGetObjectByHeapObjectIDArgs initializes GetObjectByHeapObjectIDArgs with the required arguments.
func NewGetObjectByHeapObjectIDArgs(objectID HeapSnapshotObjectID) *GetObjectByHeapObjectIDArgs {
	args := new(GetObjectByHeapObjectIDArgs)
	args.ObjectID = objectID
	return args
}

// SetObjectGroup sets the ObjectGroup optional argument. Symbolic group name that can be used to release multiple objects.
func (a *GetObjectByHeapObjectIDArgs) SetObjectGroup(objectGroup string) *GetObjectByHeapObjectIDArgs {
	a.ObjectGroup = &objectGroup
	return a
}

// GetObjectByHeapObjectIDReply represents the return values for GetObjectByHeapObjectID in the HeapProfiler domain.
type GetObjectByHeapObjectIDReply struct {
	Result runtime.RemoteObject `json:"result"` // Evaluation result.
}

// AddInspectedHeapObjectArgs represents the arguments for AddInspectedHeapObject in the HeapProfiler domain.
type AddInspectedHeapObjectArgs struct {
	HeapObjectID HeapSnapshotObjectID `json:"heapObjectId"` // Heap snapshot object id to be accessible by means of $x command line API.
}

// NewAddInspectedHeapObjectArgs initializes AddInspectedHeapObjectArgs with the required arguments.
func NewAddInspectedHeapObjectArgs(heapObjectID HeapSnapshotObjectID) *AddInspectedHeapObjectArgs {
	args := new(AddInspectedHeapObjectArgs)
	args.HeapObjectID = heapObjectID
	return args
}

// GetHeapObjectIDArgs represents the arguments for GetHeapObjectID in the HeapProfiler domain.
type GetHeapObjectIDArgs struct {
	ObjectID runtime.RemoteObjectID `json:"objectId"` // Identifier of the object to get heap object id for.
}

// NewGetHeapObjectIDArgs initializes GetHeapObjectIDArgs with the required arguments.
func NewGetHeapObjectIDArgs(objectID runtime.RemoteObjectID) *GetHeapObjectIDArgs {
	args := new(GetHeapObjectIDArgs)
	args.ObjectID = objectID
	return args
}

// GetHeapObjectIDReply represents the return values for GetHeapObjectID in the HeapProfiler domain.
type GetHeapObjectIDReply struct {
	HeapSnapshotObjectID HeapSnapshotObjectID `json:"heapSnapshotObjectId"` // Id of the heap snapshot object corresponding to the passed remote object id.
}

// StartSamplingArgs represents the arguments for StartSampling in the HeapProfiler domain.
type StartSamplingArgs struct {
	SamplingInterval *float64 `json:"samplingInterval,omitempty"` // Average sample interval in bytes. Poisson distribution is used for the intervals. The default value is 32768 bytes.
}

// NewStartSamplingArgs initializes StartSamplingArgs with the required arguments.
func NewStartSamplingArgs() *StartSamplingArgs {
	args := new(StartSamplingArgs)

	return args
}

// SetSamplingInterval sets the SamplingInterval optional argument. Average sample interval in bytes. Poisson distribution is used for the intervals. The default value is 32768 bytes.
func (a *StartSamplingArgs) SetSamplingInterval(samplingInterval float64) *StartSamplingArgs {
	a.SamplingInterval = &samplingInterval
	return a
}

// StopSamplingReply represents the return values for StopSampling in the HeapProfiler domain.
type StopSamplingReply struct {
	Profile SamplingHeapProfile `json:"profile"` // Recorded sampling heap profile.
}
