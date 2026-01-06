package utils

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	ocmConfig "github.com/openshift-online/ocm-common/pkg/ocm/config"
	sdk "github.com/openshift-online/ocm-sdk-go"
)

func resetEnvVars(t *testing.T) {
	errUrl := os.Unsetenv("OCM_URL")
	if errUrl != nil {
		t.Fatal("Error setting environment variables")
	}
}

func TestGenerateQuery(t *testing.T) {
	tests := []struct {
		name              string
		clusterIdentifier string
		want              string
	}{
		{
			name:              "valid internal ID",
			clusterIdentifier: "261kalm3uob0vegg1c7h9o7r5k9t64ji",
			want:              "(id = '261kalm3uob0vegg1c7h9o7r5k9t64ji')",
		},
		{
			name:              "valid wrong internal ID with upper case",
			clusterIdentifier: "261kalm3uob0vegg1c7h9o7r5k9t64jI",
			want:              "(display_name like '261kalm3uob0vegg1c7h9o7r5k9t64jI')",
		},
		{
			name:              "valid wrong internal ID too short",
			clusterIdentifier: "261kalm3uob0vegg1c7h9o7r5k9t64j",
			want:              "(display_name like '261kalm3uob0vegg1c7h9o7r5k9t64j')",
		},
		{
			name:              "valid wrong internal ID too long",
			clusterIdentifier: "261kalm3uob0vegg1c7h9o7r5k9t64jix",
			want:              "(display_name like '261kalm3uob0vegg1c7h9o7r5k9t64jix')",
		},
		{
			name:              "valid external ID",
			clusterIdentifier: "c1f562af-fb22-42c5-aa07-6848e1eeee9c",
			want:              "(external_id = 'c1f562af-fb22-42c5-aa07-6848e1eeee9c')",
		},
		{
			name:              "valid display name",
			clusterIdentifier: "hs-mc-773jpgko0",
			want:              "(display_name like 'hs-mc-773jpgko0')",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GenerateQuery(tt.clusterIdentifier); got != tt.want {
				t.Errorf("GenerateQuery(%s) = %v, want %v", tt.clusterIdentifier, got, tt.want)
			}
		})
	}
}

// TestGetOcmConfigFromFilePath tests the GetOcmConfigFromFilePath function which loads
// OCM configuration from a JSON file at the provided path. It validates that the function
// correctly handles valid config files, non-existent files, empty files, and malformed JSON.
func TestGetOcmConfigFromFilePath(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(t *testing.T) string
		wantErr     bool
		errContains string
	}{
		{
			// Test that a valid OCM config file is successfully parsed and loaded
			name: "valid config file",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configFile := filepath.Join(tmpDir, "ocm.json")
				config := ocmConfig.Config{
					AccessToken:  "test-access-token",
					RefreshToken: "test-refresh-token",
					URL:          "https://api.openshift.com",
					ClientID:     "test-client-id",
					ClientSecret: "test-client-secret",
				}
				data, err := json.Marshal(config)
				if err != nil {
					t.Fatalf("failed to marshal config: %v", err)
				}
				if err := os.WriteFile(configFile, data, 0644); err != nil {
					t.Fatalf("failed to write config file: %v", err)
				}
				return configFile
			},
			wantErr: false,
		},
		{
			// Test that attempting to load a non-existent file returns an appropriate error
			name: "non-existent file",
			setupFunc: func(t *testing.T) string {
				return "/nonexistent/path/ocm.json"
			},
			wantErr:     true,
			errContains: "can't read config file",
		},
		{
			// Test that an empty config file returns an error
			name: "empty config file",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configFile := filepath.Join(tmpDir, "ocm.json")
				if err := os.WriteFile(configFile, []byte(""), 0644); err != nil {
					t.Fatalf("failed to write empty file: %v", err)
				}
				return configFile
			},
			wantErr:     true,
			errContains: "empty config file",
		},
		{
			// Test that a file with invalid JSON syntax returns a parse error
			name: "invalid json",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configFile := filepath.Join(tmpDir, "ocm.json")
				if err := os.WriteFile(configFile, []byte("{invalid json}"), 0644); err != nil {
					t.Fatalf("failed to write invalid json: %v", err)
				}
				return configFile
			},
			wantErr:     true,
			errContains: "can't parse config file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := tt.setupFunc(t)
			cfg, err := GetOcmConfigFromFilePath(filePath)
			if tt.wantErr {
				if err == nil {
					t.Errorf("GetOcmConfigFromFilePath() expected error but got none")
				} else if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("GetOcmConfigFromFilePath() error = %v, want error containing %v", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("GetOcmConfigFromFilePath() unexpected error = %v", err)
				}
				if cfg == nil {
					t.Errorf("GetOcmConfigFromFilePath() returned nil config")
				}
			}
		})
	}
}

// TestGetOCMSdkConnBuilderFromConfig tests the GetOCMSdkConnBuilderFromConfig function
// which creates an OCM SDK connection builder from a provided OCM config object.
// It validates nil config handling and successful builder creation with valid config.
func TestGetOCMSdkConnBuilderFromConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *ocmConfig.Config
		wantErr bool
	}{
		{
			// Test that a valid OCM config successfully creates a connection builder
			name: "valid config",
			config: &ocmConfig.Config{
				AccessToken:  "test-access-token",
				RefreshToken: "test-refresh-token",
				URL:          "https://api.openshift.com",
				ClientID:     "test-client-id",
				ClientSecret: "test-client-secret",
			},
			wantErr: false,
		},
		{
			// Test that passing a nil config returns an error
			name:    "nil config",
			config:  nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder, err := GetOCMSdkConnBuilderFromConfig(tt.config)
			if tt.wantErr {
				if err == nil {
					t.Errorf("GetOCMSdkConnBuilderFromConfig() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("GetOCMSdkConnBuilderFromConfig() unexpected error = %v", err)
				}
				if builder == nil {
					t.Errorf("GetOCMSdkConnBuilderFromConfig() returned nil builder")
				}
			}
		})
	}
}

// TestGetOCMSdkConnBuilderFromFilePath tests the GetOCMSdkConnBuilderFromFilePath function
// which reads an OCM config file and creates an SDK connection builder from it.
// It validates both successful builder creation and error handling for invalid file paths.
func TestGetOCMSdkConnBuilderFromFilePath(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(t *testing.T) string
		wantErr     bool
		errContains string
	}{
		{
			// Test that a valid config file successfully creates a connection builder
			name: "valid config file",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configFile := filepath.Join(tmpDir, "ocm.json")
				config := ocmConfig.Config{
					AccessToken:  "test-access-token",
					RefreshToken: "test-refresh-token",
					URL:          "https://api.openshift.com",
					ClientID:     "test-client-id",
					ClientSecret: "test-client-secret",
				}
				data, err := json.Marshal(config)
				if err != nil {
					t.Fatalf("failed to marshal config: %v", err)
				}
				if err := os.WriteFile(configFile, data, 0644); err != nil {
					t.Fatalf("failed to write config file: %v", err)
				}
				return configFile
			},
			wantErr: false,
		},
		{
			// Test that attempting to load from a non-existent file returns an error
			name: "non-existent file",
			setupFunc: func(t *testing.T) string {
				return "/nonexistent/path/ocm.json"
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := tt.setupFunc(t)
			builder, err := GetOCMSdkConnBuilderFromFilePath(filePath)
			if tt.wantErr {
				if err == nil {
					t.Errorf("GetOCMSdkConnBuilderFromFilePath() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("GetOCMSdkConnBuilderFromFilePath() unexpected error = %v", err)
				}
				if builder == nil {
					t.Errorf("GetOCMSdkConnBuilderFromFilePath() returned nil builder")
				}
			}
		})
	}
}

// TestGetOCMSdkConnFromFilePath tests the GetOCMSdkConnFromFilePath function which
// reads an OCM config file and creates a fully initialized OCM SDK connection from it.
// It validates error handling for non-existent files and empty config files.
func TestGetOCMSdkConnFromFilePath(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(t *testing.T) string
		wantErr   bool
	}{
		{
			// Test that attempting to create a connection from a non-existent file returns an error
			name: "non-existent file",
			setupFunc: func(t *testing.T) string {
				return "/nonexistent/path/ocm.json"
			},
			wantErr: true,
		},
		{
			// Test that an empty config file returns an error when trying to build a connection
			name: "empty config file",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configFile := filepath.Join(tmpDir, "ocm.json")
				if err := os.WriteFile(configFile, []byte(""), 0644); err != nil {
					t.Fatalf("failed to write empty file: %v", err)
				}
				return configFile
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := tt.setupFunc(t)
			conn, err := GetOCMSdkConnFromFilePath(filePath)
			if tt.wantErr {
				if err == nil {
					t.Errorf("GetOCMSdkConnFromFilePath() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("GetOCMSdkConnFromFilePath() unexpected error = %v", err)
				}
				if conn != nil {
					defer conn.Close()
				}
			}
		})
	}
}

// TestGetHiveShardWithConn tests the GetHiveShardWithConn function which retrieves
// the hive shard URL for a cluster using a provided OCM SDK connection.
// It validates that the function properly handles nil connection inputs.
func TestGetHiveShardWithConn(t *testing.T) {
	tests := []struct {
		name      string
		clusterID string
		conn      *sdk.Connection
		wantErr   bool
	}{
		{
			// Test that passing a nil OCM connection returns an error
			name:      "nil connection",
			clusterID: "test-cluster-id",
			conn:      nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetHiveShardWithConn(tt.clusterID, tt.conn)
			if tt.wantErr {
				if err == nil {
					t.Errorf("GetHiveShardWithConn() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("GetHiveShardWithConn() unexpected error = %v", err)
				}
			}
		})
	}
}

// TestGetHiveClusterWithConn tests the GetHiveClusterWithConn function which fetches
// the hive cluster information using separate OCM connections for the target cluster
// and hive cluster. It validates the function's ability to create temporary connections
// when nil connections are provided.
func TestGetHiveClusterWithConn(t *testing.T) {
	tests := []struct {
		name       string
		clusterID  string
		clusterOCM *sdk.Connection
		hiveOCM    *sdk.Connection
		wantErr    bool
	}{
		{
			// Test that when both connections are nil, the function attempts to create a temporary connection
			// This will fail without proper OCM environment variables set
			name:       "both connections nil - should create temporary connection",
			clusterID:  "test-cluster-id",
			clusterOCM: nil,
			hiveOCM:    nil,
			wantErr:    true, // will fail when trying to create connection without proper env vars
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetHiveClusterWithConn(tt.clusterID, tt.clusterOCM, tt.hiveOCM)
			if tt.wantErr {
				if err == nil {
					t.Errorf("GetHiveClusterWithConn() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("GetHiveClusterWithConn() unexpected error = %v", err)
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && stringContains(s, substr)))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
