package local

import (
	"errors"
	"rpc/internal/flags"
	"rpc/pkg/utils"
	"testing"

	"github.com/open-amt-cloud-toolkit/go-wsman-messages/v2/pkg/wsman/amt/publickey"
	"github.com/stretchr/testify/assert"
)

var MockPRSuccess = new(MockPasswordReaderSuccess)
var MockPRFail = new(MockPasswordReaderFail)

type MockPasswordReaderSuccess struct{}

func (mpr *MockPasswordReaderSuccess) ReadPassword() (string, error) {
	return utils.TestPassword, nil
}

type MockPasswordReaderFail struct{}

func (mpr *MockPasswordReaderFail) ReadPassword() (string, error) {
	return "", errors.New("Read password failed")
}

func TestDisplayAMTInfo(t *testing.T) {
	defaultFlags := flags.AmtInfoFlags{
		Ver:      true,
		Bld:      true,
		Sku:      true,
		UUID:     true,
		Mode:     true,
		DNS:      true,
		Ras:      true,
		Lan:      true,
		Hostname: true,
		OpState:  true,
	}

	t.Run("returns Success on happy path", func(t *testing.T) {
		f := flags.NewFlags(nil, MockPRSuccess)
		f.AmtInfo = defaultFlags
		lps := setupService(f)
		err := lps.DisplayAMTInfo()
		assert.NoError(t, err)
		assert.Equal(t, nil, err)
	})

	t.Run("returns Success with json output", func(t *testing.T) {
		f := flags.NewFlags(nil, MockPRSuccess)
		f.AmtInfo = defaultFlags
		f.JsonOutput = true
		lps := setupService(f)
		err := lps.DisplayAMTInfo()
		assert.NoError(t, err)
		assert.Equal(t, nil, err)
	})

	t.Run("returns Success with certs", func(t *testing.T) {
		f := flags.NewFlags(nil, MockPRSuccess)
		f.AmtInfo.Cert = true
		f.AmtInfo.UserCert = true
		f.Password = "testPassword"
		mockCertHashes = mockCertHashesDefault
		pullEnvelope := publickey.PullResponse{}
		pullEnvelope.PublicKeyCertificateItems = []publickey.PublicKeyCertificateResponse{
			mpsCert,
			clientCert,
			caCert,
		}
		lps := setupService(f)
		err := lps.DisplayAMTInfo()
		assert.NoError(t, err)
		assert.Equal(t, nil, err)
	})

	t.Run("returns Success but logs errors on error conditions", func(t *testing.T) {
		mockUUIDErr = errMockStandard
		mockVersionDataErr = errMockStandard
		mockControlModeErr = errMockStandard
		mockDNSSuffixErr = errMockStandard
		mockOSDNSSuffixErr = errMockStandard
		mockRemoteAcessConnectionStatusErr = errMockStandard
		mockLANInterfaceSettingsErr = errMockStandard
		mockCertHashesErr = errMockStandard

		f := flags.NewFlags(nil, MockPRSuccess)
		f.AmtInfo = defaultFlags
		f.JsonOutput = true

		lps := setupService(f)
		err := lps.DisplayAMTInfo()
		assert.NoError(t, err)
		assert.Equal(t, nil, err)
		f.JsonOutput = false

		mockUUIDErr = nil
		mockVersionDataErr = nil
		mockControlModeErr = nil
		mockDNSSuffixErr = nil
		mockOSDNSSuffixErr = nil
		mockRemoteAcessConnectionStatusErr = nil
		mockLANInterfaceSettingsErr = nil
		mockCertHashesErr = nil
	})

	t.Run("resets UserCert on GetControlMode failure", func(t *testing.T) {
		f := flags.NewFlags(nil, MockPRSuccess)
		f.AmtInfo.UserCert = true
		mockControlModeErr = errMockStandard
		lps := setupService(f)
		err := lps.DisplayAMTInfo()
		assert.Equal(t, nil, err)
		assert.False(t, f.AmtInfo.UserCert)
		mockControlModeErr = nil
	})
	t.Run("resets UserCert when control mode is preprovisioning", func(t *testing.T) {
		f := flags.NewFlags(nil, MockPRSuccess)
		f.AmtInfo.UserCert = true
		orig := mockControlMode
		mockControlMode = 0
		lps := setupService(f)
		err := lps.DisplayAMTInfo()
		assert.Equal(t, nil, err)
		assert.False(t, f.AmtInfo.UserCert)
		mockControlMode = orig
	})
	t.Run("returns MissingOrIncorrectPassword on no password input from user", func(t *testing.T) {
		f := flags.NewFlags(nil, MockPRFail)
		f.AmtInfo.UserCert = true
		orig := mockControlMode
		mockControlMode = 2
		lps := setupService(f)
		err := lps.DisplayAMTInfo()
		assert.Equal(t, utils.MissingOrIncorrectPassword, err)
		assert.True(t, f.AmtInfo.UserCert)
		mockControlMode = orig
	})
}

func TestDecodeAMT(t *testing.T) {
	testCases := []struct {
		version string
		SKU     string
		want    string
	}{
		{"200", "0", "Invalid AMT version format"},
		{"ab.c", "0", "Invalid AMT version"},
		{"2.0.0", "0", "AMT + ASF + iQST"},
		{"2.1.0", "1", "ASF + iQST"},
		{"2.2.0", "2", "iQST"},
		{"1.1.0", "3", "Unknown"},
		{"3.0.0", "008", "Invalid SKU"},
		{"3.0.0", "8", "AMT"},
		{"4.1.0", "2", "iQST "},
		{"4.0.0", "4", "ASF "},
		{"5.0.0", "288", "TPM Home IT "},
		{"5.0.0", "1088", "WOX "},
		{"5.0.0", "38", "iQST ASF TPM "},
		{"5.0.0", "4", "ASF "},
		{"6.0.0", "2", "iQST "},
		{"7.0.0", "36864", "L3 Mgt Upgrade"},
		{"8.0.0", "24584", "AMT Pro AT-p Corporate "},
		{"10.0.0", "8", "AMT Pro "},
		{"11.0.0", "16392", "AMT Pro Corporate "},
		{"15.0.42", "16392", "AMT Pro Corporate "},
		{"16.1.25", "16400", "Intel Standard Manageability Corporate "},
	}

	for _, tc := range testCases {
		got := DecodeAMT(tc.version, tc.SKU)
		if got != tc.want {
			t.Errorf("DecodeAMT(%q, %q) = %v; want %v", tc.version, tc.SKU, got, tc.want)
		}
	}
}
