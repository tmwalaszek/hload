package templates

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmwalaszek/hload/model"
)

func TestRenderOutput(t *testing.T) {
	loaders := &Loaders{
		Loaders: []LoaderSummaries{
			{
				Loader: &model.Loader{
					Name: "name",
				},
			},
		},
	}

	r, err := NewRenderTemplate("default", "")
	require.NoError(t, err)
	_, err = r.RenderOutput(loaders)
	require.NoError(t, err)
}
