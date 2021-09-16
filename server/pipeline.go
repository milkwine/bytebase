package server

import (
	"context"

	"github.com/bytebase/bytebase/api"
)

func (s *Server) ComposePipelineById(ctx context.Context, id int) (*api.Pipeline, error) {
	pipelineFind := &api.PipelineFind{
		ID: &id,
	}
	pipeline, err := s.PipelineService.FindPipeline(context.Background(), pipelineFind)
	if err != nil {
		return nil, err
	}

	if err := s.ComposePipelineRelationship(ctx, pipeline); err != nil {
		return nil, err
	}

	return pipeline, nil
}

func (s *Server) ComposePipelineRelationship(ctx context.Context, pipeline *api.Pipeline) error {
	var err error

	pipeline.Creator, err = s.ComposePrincipalById(context.Background(), pipeline.CreatorId)
	if err != nil {
		return err
	}

	pipeline.Updater, err = s.ComposePrincipalById(context.Background(), pipeline.UpdaterId)
	if err != nil {
		return err
	}

	pipeline.StageList, err = s.ComposeStageListByPipelineId(context.Background(), pipeline.ID)
	if err != nil {
		return err
	}

	return nil
}

// Try to schedule the next task if needed
func (s *Server) ScheduleNextTaskIfNeeded(ctx context.Context, pipeline *api.Pipeline) error {
	for _, stage := range pipeline.StageList {
		for _, task := range stage.TaskList {
			skipIfAlreadyDone := true
			if task.Status == api.TaskPendingApproval {
				return s.TaskCheckScheduler.ScheduleCheckIfNeeded(ctx, task, api.SYSTEM_BOT_ID, skipIfAlreadyDone)
			}

			if task.Status == api.TaskPending {
				if err := s.TaskCheckScheduler.ScheduleCheckIfNeeded(ctx, task, api.SYSTEM_BOT_ID, skipIfAlreadyDone); err != nil {
					return err
				}
				_, err := s.TaskScheduler.Schedule(context.Background(), task)
				if err != nil {
					return err
				}
				return nil
			}
		}
	}
	return nil
}
