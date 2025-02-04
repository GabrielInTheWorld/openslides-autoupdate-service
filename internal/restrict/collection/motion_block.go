package collection

import (
	"context"
	"fmt"

	"github.com/OpenSlides/openslides-autoupdate-service/internal/restrict/perm"
	"github.com/OpenSlides/openslides-autoupdate-service/pkg/datastore"
)

// MotionBlock handels restrictions of the collection motion_block.
type MotionBlock struct{}

// Modes returns the restrictions modes for the meeting collection.
func (m MotionBlock) Modes(mode string) FieldRestricter {
	switch mode {
	case "A":
		return m.modeA
	case "B":
		return m.see
	}
	return nil
}

func (m MotionBlock) see(ctx context.Context, fetch *datastore.Fetcher, mperms *perm.MeetingPermission, motionBlockID int) (bool, error) {
	meetingID := fetch.Field().MotionBlock_MeetingID(ctx, motionBlockID)
	if err := fetch.Err(); err != nil {
		return false, fmt.Errorf("getting meetingID: %w", err)
	}

	perms, err := mperms.Meeting(ctx, meetingID)
	if err != nil {
		return false, fmt.Errorf("getting permissions: %w", err)
	}

	if perms.Has(perm.MotionCanManage) {
		return true, nil
	}

	if !perms.Has(perm.MotionCanSee) {
		return false, nil
	}

	internal := fetch.Field().MotionBlock_Internal(ctx, motionBlockID)
	if err := fetch.Err(); err != nil {
		return false, fmt.Errorf("getting internal: %w", err)
	}

	if !internal {
		return true, nil
	}

	return false, nil
}

func (m MotionBlock) modeA(ctx context.Context, fetch *datastore.Fetcher, mperms *perm.MeetingPermission, motionBlockID int) (bool, error) {
	see, err := m.see(ctx, fetch, mperms, motionBlockID)
	if err != nil {
		return false, fmt.Errorf("checking see: %w", err)
	}

	if see {
		return true, nil
	}

	agendaItemID := fetch.Field().MotionBlock_AgendaItemID(ctx, motionBlockID)
	if err := fetch.Err(); err != nil {
		return false, fmt.Errorf("getting agendaItem: %w", err)
	}

	if agendaItemID != 0 {
		see, err = AgendaItem{}.see(ctx, fetch, mperms, agendaItemID)
		if err != nil {
			return false, fmt.Errorf("checking agendaItem %d: %w", agendaItemID, err)
		}

		if see {
			return true, nil
		}
	}

	losID := fetch.Field().MotionBlock_ListOfSpeakersID(ctx, motionBlockID)
	if err := fetch.Err(); err != nil {
		return false, fmt.Errorf("getting list of speakers: %w", err)
	}

	if losID != 0 {
		see, err = ListOfSpeakers{}.see(ctx, fetch, mperms, losID)
		if err != nil {
			return false, fmt.Errorf("checking list of speakers %d: %w", losID, err)
		}

		if see {
			return true, nil
		}
	}
	return false, nil
}
