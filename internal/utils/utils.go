package utils

import (
	"encoding/base64"
	"fmt"

	"github.com/zuczkows/text-bot-integration/internal/config"
)

func GenerateBasicAuthToken(accountID string, personalToken config.Secret) string {
	credentials := fmt.Sprintf("%s:%s", accountID, personalToken)
	encoded := base64.StdEncoding.EncodeToString([]byte(credentials))
	return fmt.Sprintf("Basic %s", encoded)
}
