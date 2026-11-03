package setup

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"styx/internal/db"
	"time"

	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/google/uuid"
)

func GetMachineID() (string, error) {
	id, err := os.ReadFile("/etc/machine-id")
	return string(id), err
}

func GetProductID() (string, error) {
	id, err := os.ReadFile("/sys/class/dmi/id/product_uuid")
	return string(id), err
}
func GetProductSERIAL() (string, error) {
	id, err := os.ReadFile("/sys/class/dmi/id/product_serial")
	return string(id), err
}

func CollectRegitrationInfo(user, pass string) (db.LicenseInfo, error) {
	saltBytes := make([]byte, 32)
	if _, err := rand.Read(saltBytes); err != nil {
		return db.LicenseInfo{}, fmt.Errorf("generate salt: %w", err)
	}
	salt := hex.EncodeToString(saltBytes)
	hash := func(value string) string {
		sum := sha256.Sum256([]byte(salt + ":" + value))
		return hex.EncodeToString(sum[:])
	}
	id, err := uuid.NewV7()
	if err != nil {
		return db.LicenseInfo{}, err
	}
	licenseID := id.String()

	MachineId, err := GetMachineID()
	if err != nil {
		return db.LicenseInfo{}, err
	}

	ProductID, err := GetProductID()
	if err != nil {
		return db.LicenseInfo{}, err
	}

	SerialID, err := GetProductSERIAL()
	if err != nil {
		return db.LicenseInfo{}, err
	}

	serialHash := hash(SerialID)
	machineIDhash := hash(MachineId)
	ProductIDhash := hash(ProductID)
	LicenseHash := hash(licenseID)
	MetaHash := hash(fmt.Sprintf("%s:%s:%s:%s", LicenseHash, machineIDhash, ProductIDhash, serialHash))
	PassHash := hash(pass)

	return db.LicenseInfo{MetaHash: MetaHash, Salt: salt,
		LicenseID: LicenseHash, MachineID: machineIDhash,
		SerialID: serialHash, ProductID: ProductIDhash, User: user,
		Pass: PassHash}, nil
}

type RegisterReq struct {
	User      string `json:"user"`
	Pass      string `json:"pass"`
	MachineId string `json:"machine_id"`
	LicenseId string `json:"license_id"`
	ProductID string `json:"product_id"`
	SerialId  string `json:"serial_id"`
	Salt      string `json:"salt"`
}

type RegisterResp struct {
	UserID   string `json:"user_id"`
	ErrorMsg string `json:"error"`
}

func RegisterToServer(domain string, info db.LicenseInfo) (string, error) {
	if !strings.Contains(domain, "://") {
		domain = "https://" + domain
	}
	u, err := url.Parse(domain)
	if err != nil || u.Hostname() == "" || u.User != nil || (u.Path != "" && u.Path != "/") || u.RawQuery != "" || u.Fragment != "" {
		return "", fmt.Errorf("invalid registration server %q", domain)
	}
	ip := net.ParseIP(u.Hostname())
	if u.Scheme != "https" && !(u.Scheme == "http" && (u.Hostname() == "localhost" || ip != nil && ip.IsLoopback())) {
		return "", fmt.Errorf("registration server must use HTTPS, except for localhost")
	}
	u.Path = "/rest/user/register"
	backend := u.String()
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	body := RegisterReq{
		User:      info.User,
		Pass:      info.Pass,
		MachineId: info.MachineID,
		ProductID: info.ProductID,
		LicenseId: info.LicenseID,
		SerialId:  info.SerialID,
		Salt:      info.Salt,
	}

	data, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, backend, bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var apiresp RegisterResp
	if err := json.NewDecoder(resp.Body).Decode(&apiresp); err != nil {
		return "", err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("registration failed (%s): %s", resp.Status, apiresp.ErrorMsg)
	}
	if apiresp.UserID == "" {
		return "", fmt.Errorf("registration response is missing user_id")
	}

	return apiresp.UserID, nil

}
