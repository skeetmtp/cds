package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/ovh/cds/engine/api/cache"
	"github.com/ovh/cds/engine/api/project"
	"github.com/ovh/cds/engine/api/workflow"
	"github.com/ovh/cds/engine/service"
	"github.com/ovh/cds/sdk"
)

func (api *API) getWorkflowAsCodeHandler() service.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		vars := mux.Vars(r)
		uuid := vars["uuid"]

		var ope sdk.Operation
		b := api.Cache.Get(cache.Key(workflow.CacheOperationKey, uuid), &ope)
		if !b {
			return sdk.ErrNotFound
		}
		return service.WriteJSON(w, ope, http.StatusOK)
	}
}

// postWorkflowAsCodeHandler Make the workflow as code
// @title Make the workflow as code
func (api *API) postWorkflowAsCodeHandler() service.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		vars := mux.Vars(r)
		key := vars["key"]
		workflowName := vars["permWorkflowName"]

		u := getUser(ctx)

		proj, errP := project.Load(api.mustDB(), api.Cache, key, u, project.LoadOptions.WithApplicationWithDeploymentStrategies, project.LoadOptions.WithPipelines, project.LoadOptions.WithEnvironments, project.LoadOptions.WithPlatforms)
		if errP != nil {
			return sdk.WrapError(errP, "unable to load project")
		}
		wf, errW := workflow.Load(ctx, api.mustDB(), api.Cache, proj, workflowName, u, workflow.LoadOptions{
			DeepPipeline:          true,
			WithAsCodeUpdateEvent: true,
		})
		if errW != nil {
			return sdk.WrapError(errW, "unable to load workflow")
		}

		// Workflow must have a repository webhook on root node
		found := false
		for _, h := range wf.WorkflowData.Node.Hooks {
			if h.HookModelName == sdk.RepositoryWebHookModel.Name {
				found = true
				break
			}
		}
		if !found {
			return sdk.NewErrorFrom(sdk.ErrWrongRequest, "Root node must have a repository web hook")
		}

		// Sync as code event
		if len(wf.AsCodeEvent) > 0 {
			tx, err := api.mustDB().Begin()
			if err != nil {
				return sdk.WrapError(err, "unable to start transaction")
			}
			if err := workflow.SyncAsCodeEvent(ctx, tx, api.Cache, proj, wf); err != nil {
				tx.Rollback() // nolint
				return err
			}
			if err := tx.Commit(); err != nil {
				return sdk.WrapError(err, "unable to commit transaction")
			}
		}

		if wf.FromRepository != "" || (wf.FromRepository == "" && len(wf.AsCodeEvent) > 0) {
			return sdk.WithStack(sdk.ErrWorkflowAlreadAsCode)
		}

		ope, err := workflow.UpdateAsCode(ctx, api.mustDB(), api.Cache, proj, wf, project.EncryptWithBuiltinKey, u)
		if err != nil {
			return sdk.WrapError(errW, "unable to migrate workflow as code")
		}

		sdk.GoRoutine(context.Background(), fmt.Sprintf("UpdateWorkflowAsCodeResult-%s", ope.UUID), func(ctx context.Context) {
			workflow.UpdateWorkflowAsCodeResult(ctx, api.mustDB(), api.Cache, proj, ope, wf, u)
		}, api.PanicDump())

		return service.WriteJSON(w, ope, http.StatusOK)
	}
}
