package collection

import (
	"context"
	"fmt"

	"github.com/OpenSlides/openslides-autoupdate-service/internal/restrict/perm"
	"github.com/OpenSlides/openslides-autoupdate-service/pkg/datastore"
)

// Committee handels permission for committees.
type Committee struct{}

// Modes returns a map from all known modes to there restricter.
func (a Committee) Modes(mode string) FieldRestricter {
	switch mode {
	case "A":
		return a.see
	case "B":
		return a.modeB
	}
	return nil
}

func (a Committee) see(ctx context.Context, fetch *datastore.Fetcher, mperms *perm.MeetingPermission, committeeID int) (bool, error) {
	userIDs := fetch.Field().Committee_UserIDs(ctx, committeeID)
	if err := fetch.Err(); err != nil {
		return false, fmt.Errorf("getting committee users: %w", err)
	}

	for _, uid := range userIDs {
		if uid == mperms.UserID() {
			return true, nil
		}
	}

	hasOMLPerm, err := perm.HasOrganizationManagementLevel(ctx, fetch, mperms.UserID(), perm.OMLCanManageUsers)
	if err != nil {
		return false, fmt.Errorf("checking oml perm: %w", err)
	}

	return hasOMLPerm, nil
}

func (a Committee) modeB(ctx context.Context, fetch *datastore.Fetcher, mperms *perm.MeetingPermission, committeeID int) (bool, error) {
	hasOMLPerm, err := perm.HasOrganizationManagementLevel(ctx, fetch, mperms.UserID(), perm.OMLCanManageOrganization)
	if err != nil {
		return false, fmt.Errorf("checking oml: %w", err)
	}
	return hasOMLPerm, nil
}
