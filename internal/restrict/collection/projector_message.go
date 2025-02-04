package collection

import (
	"context"
	"fmt"

	"github.com/OpenSlides/openslides-autoupdate-service/internal/restrict/perm"
	"github.com/OpenSlides/openslides-autoupdate-service/pkg/datastore"
)

// ProjectorMessage handels the restriction for the projector_message collection.
type ProjectorMessage struct{}

// Modes returns the restrictions modes for the meeting collection.
func (p ProjectorMessage) Modes(mode string) FieldRestricter {
	switch mode {
	case "A":
		return p.see
	}
	return nil
}

func (p ProjectorMessage) see(ctx context.Context, fetch *datastore.Fetcher, mperms *perm.MeetingPermission, projectorMessageID int) (bool, error) {
	meetingID := fetch.Field().ProjectorMessage_ID(ctx, projectorMessageID)
	if err := fetch.Err(); err != nil {
		return false, fmt.Errorf("fetching meeting_id %d: %w", projectorMessageID, err)
	}

	perms, err := mperms.Meeting(ctx, meetingID)
	if err != nil {
		return false, fmt.Errorf("getting perms for meeting %d: %w", meetingID, err)
	}

	return perms.Has(perm.ProjectorCanSee), nil
}
