package contrailclient

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHasRequiredAnnotations(t *testing.T) {
	tests := []struct{
		name string
		actualAnnotations map[string]string
		requiredAnnotations map[string]string
		expectedResult bool
	}{
		{
			name: "should return true when required annotations are equal to actual annotations",
			actualAnnotations: map[string]string{
				"test": "annotation",
			},
			requiredAnnotations: map[string]string{
				"test": "annotation",
			},
			expectedResult: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actualResult := HasRequiredAnnotations(test.actualAnnotations, test.requiredAnnotations)
			assert.Equal(t, actualResult, test.expectedResult)
		})
	} 
}