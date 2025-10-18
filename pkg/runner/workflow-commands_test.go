package runner

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// Based on github.com/actions/runner/src/Test/L0/Worker/ActionCommandL0.cs

func TestParseCommandV2Cases(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantOK      bool
		wantCommand WorkflowCommandName
		wantData    string
		wantProps   map[string]string
	}{
		{
			name:        "SimpleCommand",
			input:       "::debug k1=v1,::msg",
			wantOK:      true,
			wantCommand: WorkflowCommandNameDebug,
			wantData:    "msg",
			wantProps: map[string]string{
				"k1": "v1",
			},
		},
		{
			name:        "EmptyData",
			input:       "::debug::",
			wantOK:      true,
			wantCommand: WorkflowCommandNameDebug,
		},
		{
			name:        "EscapedPropertiesAndData",
			input:       "::debug k1=;=%2C=%0D=%0A=]=%3A,::;-%0D-%0A-]-:-,",
			wantOK:      true,
			wantCommand: WorkflowCommandNameDebug,
			wantData:    ";-\r-\n-]-:-,",
			wantProps: map[string]string{
				"k1": ";=,=\r=\n=]=:",
			},
		},
		{
			name:        "DoubleEscaped",
			input:       "::debug k1=;=%252C=%250D=%250A=]=%253A,::;-%250D-%250A-]-:-,",
			wantOK:      true,
			wantCommand: WorkflowCommandNameDebug,
			wantData:    ";-%0D-%0A-]-:-,",
			wantProps: map[string]string{
				"k1": ";=%2C=%0D=%0A=]=%3A",
			},
		},
		{
			name:        "IgnoreEmptyPropertyValues",
			input:       "::debug k1=,k2=,::",
			wantOK:      true,
			wantCommand: WorkflowCommandNameDebug,
		},
		{
			name:        "SinglePropertyNoData",
			input:       "::debug k1=v1::",
			wantOK:      true,
			wantCommand: WorkflowCommandNameDebug,
			wantProps: map[string]string{
				"k1": "v1",
			},
		},
		{
			name:        "TrimmedPrefix",
			input:       "   ::debug k1=v1,::msg",
			wantOK:      true,
			wantCommand: WorkflowCommandNameDebug,
			wantData:    "msg",
			wantProps: map[string]string{
				"k1": "v1",
			},
		},
		{
			name:   "NonCommandPrefix",
			input:  "   >>>   ::debug k1=v1,::msg",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			got, ok := parseWorkflowCommandV2(ctx, tt.input)
			require.Equal(t, tt.wantOK, ok)
			if !tt.wantOK {
				return
			}

			require.Equal(t, tt.wantCommand, got.Command)
			require.Equal(t, tt.wantData, got.Data)

			if tt.wantProps == nil {
				require.Len(t, got.Props, 0)
				return
			}

			require.NotNil(t, got.Props)
			require.Equal(t, tt.wantProps, got.Props)
		})
	}
}

func TestParseCommandV1Cases(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantOK      bool
		wantCommand WorkflowCommandName
		wantData    string
		wantProps   map[string]string
	}{
		{
			name:        "SimpleCommand",
			input:       "##[debug k1=v1;]msg",
			wantOK:      true,
			wantCommand: WorkflowCommandNameDebug,
			wantData:    "msg",
			wantProps: map[string]string{
				"k1": "v1",
			},
		},
		{
			name:        "EmptyData",
			input:       "##[debug]",
			wantOK:      true,
			wantCommand: WorkflowCommandNameDebug,
		},
		{
			name:        "EscapedPropertiesAndData",
			input:       "##[debug k1=%3B=%0D=%0A=%5D;]%3B-%0D-%0A-%5D",
			wantOK:      true,
			wantCommand: WorkflowCommandNameDebug,
			wantData:    ";-\r-\n-]",
			wantProps: map[string]string{
				"k1": ";=\r=\n=]",
			},
		},
		{
			name:        "DoubleEscaped",
			input:       "##[debug k1=%253B=%250D=%250A=%255D;]%253B-%250D-%250A-%255D",
			wantOK:      true,
			wantCommand: WorkflowCommandNameDebug,
			wantData:    "%3B-%0D-%0A-%5D",
			wantProps: map[string]string{
				"k1": "%3B=%0D=%0A=%5D",
			},
		},
		{
			name:        "IgnoreEmptyPropertyValues",
			input:       "##[debug k1=;k2=;]",
			wantOK:      true,
			wantCommand: WorkflowCommandNameDebug,
		},
		{
			name:        "PrefixedCommand",
			input:       ">>>   ##[debug k1=v1;]msg",
			wantOK:      true,
			wantCommand: WorkflowCommandNameDebug,
			wantData:    "msg",
			wantProps: map[string]string{
				"k1": "v1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			got, ok := parseWorkflowCommandV1(ctx, tt.input)
			require.Equal(t, tt.wantOK, ok)
			if !tt.wantOK {
				return
			}

			require.Equal(t, tt.wantCommand, got.Command)
			require.Equal(t, tt.wantData, got.Data)

			if tt.wantProps == nil {
				require.Len(t, got.Props, 0)
				return
			}

			require.NotNil(t, got.Props)
			require.Equal(t, tt.wantProps, got.Props)
		})
	}
}
