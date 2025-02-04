package collection_test

import (
	"testing"

	"github.com/OpenSlides/openslides-autoupdate-service/internal/restrict/collection"
	"github.com/OpenSlides/openslides-autoupdate-service/internal/restrict/perm"
)

func TestProjectorMessageModeA(t *testing.T) {
	f := collection.ProjectorMessage{}.Modes("A")

	testCase(
		"can see",
		t,
		f,
		true,
		"projector_message/1/meeting_id: 1",
		withPerms(1, perm.ProjectorCanSee),
	)

	testCase(
		"no perm",
		t,
		f,
		false,
		"projector_message/1/meeting_id: 1",
	)
}
