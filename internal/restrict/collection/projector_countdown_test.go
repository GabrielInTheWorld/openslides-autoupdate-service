package collection_test

import (
	"testing"

	"github.com/OpenSlides/openslides-autoupdate-service/internal/restrict/collection"
	"github.com/OpenSlides/openslides-autoupdate-service/internal/restrict/perm"
)

func TestProjectorCountdownModeA(t *testing.T) {
	f := collection.ProjectorCountdown{}.Modes("A")

	testCase(
		"can see",
		t,
		f,
		true,
		"projector_countdown/1/meeting_id: 1",
		withPerms(1, perm.ProjectorCanSee),
	)

	testCase(
		"no perm",
		t,
		f,
		false,
		"projector_countdown/1/meeting_id: 1",
	)
}
