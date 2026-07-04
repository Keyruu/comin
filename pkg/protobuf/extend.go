package protobuf

import "fmt"

// Short return a short line for each logs.
// This function is experimental and subject to change.
func (event *Event) Short() string {
	var eventType, payload string
	switch t := event.Type.(type) {
	case *Event_EvalStartedType:
		eventType = "eval-started   "
		g := t.EvalStartedType.Generation
		payload = fmt.Sprintf("gen-uuid=%s ref-id=%s/%s/%s", g.Uuid, g.SelectedRemoteName, g.SelectedBranchName, g.SelectedCommitId)
	case *Event_EvalFinishedType:
		eventType = "eval-finished  "
		g := t.EvalFinishedType.Generation
		payload = fmt.Sprintf("gen-uuid=%s ref-id=%s/%s/%s status=%s", g.Uuid, g.SelectedRemoteName, g.SelectedBranchName, g.SelectedCommitId, g.EvalStatus)
	case *Event_BuildStartedType:
		eventType = "build-started  "
		g := t.BuildStartedType.Generation
		payload = fmt.Sprintf("gen-uuid=%s ref-id=%s/%s/%s", g.Uuid, g.SelectedRemoteName, g.SelectedBranchName, g.SelectedCommitId)
	case *Event_BuildFinishedType:
		eventType = "build-finished "
		g := t.BuildFinishedType.Generation
		payload = fmt.Sprintf("gen-uuid=%s ref-id=%s/%s/%s status=%s", g.Uuid, g.SelectedRemoteName, g.SelectedBranchName, g.SelectedCommitId, g.BuildStatus)
	case *Event_ConfirmationSubmittedType:
		eventType = "cf-submitted   "
	case *Event_ConfirmationCancelledType:
		eventType = "cf-cancelled   "
	case *Event_ConfirmationConfirmedType:
		eventType = "cf-confirmed   "
	case *Event_Resume_:
		eventType = "resume         "
	case *Event_Suspend_:
		eventType = "suspend        "
	case *Event_DeploymentStartedType:
		eventType = "dpl-started    "
		g := t.DeploymentStartedType.Deployment.Generation
		d := t.DeploymentStartedType.Deployment
		payload = fmt.Sprintf("gen-uuid=%s ref-id=%s/%s/%s op=%s", g.Uuid, g.SelectedRemoteName, g.SelectedBranchName, g.SelectedCommitId, d.Operation)
	case *Event_DeploymentFinishedType:
		eventType = "dpl-finished   "
		g := t.DeploymentFinishedType.Deployment.Generation
		d := t.DeploymentFinishedType.Deployment
		payload = fmt.Sprintf("gen-uuid=%s ref-id=%s/%s/%s op=%s status=%s", g.Uuid, g.SelectedRemoteName, g.SelectedBranchName, g.SelectedCommitId, d.Operation, d.Status)
	case *Event_RebootRequired_:
		eventType = "reboot-required"
	case *Event_ManagerState_:
		eventType = "manager-state  "
	case *Event_Fetched_:
		eventType = "fetched        "
	case *Event_Log_:
		eventType = "log            "
		switch l := t.Log.Type.(type) {
		case *Event_Log_Open_:
			payload = "open"
		case *Event_Log_Close_:
			payload = "close"
		case *Event_Log_Line_:
			payload = fmt.Sprintf("src=%s msg=%s", l.Line.Source, l.Line.Msg)
		}
	default:
		eventType = "unknown"
	}
	return fmt.Sprintf("%s %s", eventType, payload)
}
