package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	UserFieldSubdomain = "user-field-eu-1a2039d9"
	AdsFieldSubdomain  = "ads-field-eu-1a2039d9"
	OwletAppID         = "OwletCare-Android-EU-fw-id"
	OwletAppSecret     = "OwletCare-Android-EU-JKupMPBoj_Npce_9a95Pc8Qo0Mw"
	ApiKey             = "AIzaSyDm6EhV70wudwN3iOSq3vTjtsdGjdFLuuM"
	AndroidPackage     = "com.owletcare.owletcare"
	AndroidCert        = "2A3BC26DB0B8B0792DBE28E6FFDC2598F9B12B74"
)

type Application struct {
	ID     string `json:"app_id"`
	Secret string `json:"app_secret"`
}

type User struct {
	Email       string      `json:"email"`
	Password    string      `json:"password"`
	Application Application `json:"application"`
}

type Payload struct {
	User User `json:"user"`
}

type Device struct {
	DSN              string    `json:"dsn"`
	ProductName      string    `json:"product_name"`
	Model            string    `json:"model"`
	ConnectionStatus string    `json:"connection_status"`
	DeviceType       string    `json:"device_type"`
	SWVersion        string    `json:"sw_version"`
	Mac              string    `json:"mac"`
	ConnectedAt      time.Time `json:"connected_at"`
}

type DeviceRoot struct {
	Device Device `json:"device"`
}

type Property struct {
	Key         int       `json:"key"`
	BaseType    string    `json:"base_type"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Value       FlexValue `json:"value"`
	UpdatedAt   FlexTime  `json:"data_updated_at"`
}

type FlexValue string

type FlexTime struct {
	time.Time
}

type RealTimeVitals struct {
	OxygenSaturation    int    `json:"ox"`
	HeartRate           int    `json:"hr"`
	BatteryPercentage   int    `json:"bat"`
	BatteryMinutes      int    `json:"btt"`
	SignalStrength      int    `json:"rsi"`
	OxygenTenAV         int    `json:"oxta"`
	SockConnection      int    `json:"sc"`
	SleepState          int    `json:"ss"`
	SkinTemperature     int    `json:"st"`
	Movement            int    `json:"mv"`
	AlertPausedStatus   int    `json:"aps"`
	Charging            int    `json:"chg"`
	AlertsMask          int    `json:"alrt"`
	UpdateStatus        int    `json:"ota"`
	ReadingFlags        int    `json:"srf"`
	BrickStatus         int    `json:"sb"`
	MovementBucket      int    `json:"mvb"`
	WellnessAlert       int    `json:"onm"`
	MonitoringStartTime int    `json:"mst"`
	BaseBatteryStatus   int    `json:"bsb"`
	BaseStationOn       int    `json:"bso"`
	HardwareVersion     string `json:"hw"`
}

func (fv *FlexValue) UnmarshalJSON(b []byte) error {
	if b[0] == '"' {
		return json.Unmarshal(b, (*string)(fv))
	}

	if string(b) == "null" {
		*fv = FlexValue("")
		return nil
	}

	*fv = FlexValue(fmt.Sprintf("%s", string(b)))
	return nil
}

func (ft *FlexTime) UnmarshalJSON(b []byte) error {
	if bytes.Compare(b, []byte{'"', 'n', 'u', 'l', 'l', '"'}) == 0 {
		*ft = FlexTime{}
		return nil
	}

	var currTime time.Time
	err := json.Unmarshal(b, &currTime)
	if err != nil {
		return err
	}
	*ft = FlexTime{currTime}
	return nil
}

type IntDatapoint struct {
	Value     int                  `json:"value"`
	Metadata  map[string]FlexValue `json:"metadata"`
	UpdatedAt time.Time            `json:"updated_at"`
}

type datapointRequest struct {
	Datapoint IntDatapoint `json:"datapoint"`
}

type PropertyRoot struct {
	Property *Property `json:"property"`
}

type Authentication struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Role         string `json:"role"`
}

type Client struct {
	implementationClient *http.Client
	Email                string
	Password             string
	ActivePropID         int
	TokenExpiry          time.Time
	Auth                 *Authentication
	Device               *Device
}

func (c *Client) passwordVerification() error {
	url := fmt.Sprintf("https://www.googleapis.com/identitytoolkit/v3/relyingparty/verifyPassword?key=%s", ApiKey)

	data := map[string]interface{}{
		"email":             c.Email,
		"password":          c.Password,
		"returnSecureToken": true,
	}

	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}

	dataReader := bytes.NewReader(payload)

	req, err := http.NewRequest("POST", url, dataReader)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Android-Package", AndroidPackage)
	req.Header.Set("X-Android-Cert", AndroidCert)

	resp, err := c.implementationClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	dict := map[string]interface{}{}
	err = json.NewDecoder(resp.Body).Decode(&dict)
	if err != nil {
		return fmt.Errorf("failed to decode password verify response: %w", err)
	}

	refreshToken := dict["refreshToken"]
	if c.Auth == nil {
		c.Auth = &Authentication{}
	}

	c.Auth.RefreshToken = refreshToken.(string)

	return nil

}

func (c *Client) tokenExpired() bool {
	return time.Now().After(c.TokenExpiry)
}

func (c *Client) authenticate() error {
	if c.Auth == nil {
		if c.Email == "" || c.Password == "" {
			return fmt.Errorf("email/password not supplied")
		}

		if err := c.passwordVerification(); err != nil {
			return err
		}
	}

	if c.Auth.AccessToken == "" || c.tokenExpired() {
		if err := c.refreshAuth(); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) Post(subdomain, endpoint string, data interface{}, v interface{}) error {
	return c.MakeRequest("POST", subdomain, endpoint, data, v)
}

func (c *Client) Get(subdomain, endpoint string, v interface{}) error {
	return c.MakeRequest("GET", subdomain, endpoint, nil, v)
}

func (c *Client) MakeRequest(method, subdomain, endpoint string, data interface{}, v interface{}) error {
	resp, err := c.doWithAuthorization(method, subdomain, endpoint, data, v)
	if err != nil {
		return err
	}

	if resp.StatusCode == 401 {
		// Reauthorize/Refresh and re-run request if 401.
		err = c.authenticate()
		if err != nil {
			return err
		}

		resp, err = c.doWithAuthorization(method, subdomain, endpoint, data, v)
		if err != nil {
			return err
		}
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 300 || resp.StatusCode < 200 {
		return fmt.Errorf("request failed with status: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	return json.Unmarshal(body, v)
}

func (c *Client) doWithAuthorization(method, subdomain, endpoint string, data, v interface{}) (*http.Response, error) {
	req, err := NewRequestWithAuthorization(c.Auth, method, subdomain, endpoint, data)
	if err != nil {
		return nil, err
	}

	return c.implementationClient.Do(req)
}

func (c *Client) SetFirstDevice() error {
	devices, err := c.GetDevices()
	if err != nil {
		return err
	}

	c.Device = &devices[0]
	return nil
}

func (c *Client) GetDevices() ([]Device, error) {
	deviceRoots := make([]DeviceRoot, 0)
	err := c.Get(AdsFieldSubdomain, "apiv1/devices.json", &deviceRoots)
	if err != nil {
		return []Device{}, err
	}

	devices := make([]Device, len(deviceRoots))
	for i, v := range deviceRoots {
		devices[i] = v.Device
	}
	return devices, nil
}

func (c *Client) GetRealTimeVitals(deviceID string) (*RealTimeVitals, error) {
	realTimeVitalsProp, err := c.GetPropertyByName(deviceID, "REAL_TIME_VITALS")
	if err != nil {
		return nil, fmt.Errorf("failed to get property REAL_TIME_VITALS: %w", err)
	}

	if realTimeVitalsProp == nil {
		return nil, fmt.Errorf("the property for REAL_TIME_VITALS is nil")
	}

	realTimeVitals := &RealTimeVitals{}
	if err := json.Unmarshal(json.RawMessage(realTimeVitalsProp.Value), realTimeVitals); err != nil {
		return nil, fmt.Errorf("failed to unmarshal value from REAL_TIME_VITALS property: %w", err)
	}

	return realTimeVitals, nil
}

func (c *Client) GetPropertyByName(deviceID, name string) (*Property, error) {
	endpoint := fmt.Sprintf("apiv1/dsns/%s/properties/%s", deviceID, name)
	propertyRoot := &PropertyRoot{}
	err := c.Get(AdsFieldSubdomain, endpoint, propertyRoot)
	return propertyRoot.Property, err
}

func (c *Client) GetProperties(deviceID string) (map[string]*Property, error) {
	endpoint := fmt.Sprintf("apiv1/dsns/%s/properties.json", deviceID)

	propertyRoots := make([]PropertyRoot, 0)
	err := c.Get(AdsFieldSubdomain, endpoint, &propertyRoots)
	if err != nil {
		return make(map[string]*Property), err
	}
	properties := make(map[string]*Property)
	for _, v := range propertyRoots {
		property := v.Property
		properties[v.Property.Name] = property
	}

	return properties, nil
}

func (c *Client) SetAppActiveStatus(deviceID string) (bool, error) {
	endpoint := fmt.Sprintf("apiv1/dsns/%s/properties/APP_ACTIVE/datapoints.json", deviceID)

	reqDP := datapointRequest{
		Datapoint: IntDatapoint{
			Value: 1,
		},
	}

	respDP := &datapointRequest{}
	if err := c.Post(AdsFieldSubdomain, endpoint, reqDP, respDP); err != nil {
		return false, err
	}

	return true, nil
}

func New(email, password string) (*Client, error) {
	c := &Client{
		Email:                email,
		Password:             password,
		implementationClient: http.DefaultClient,
	}

	err := c.authenticate()
	return c, err
}

func NewRequestWithAuthorization(auth *Authentication, method, subdomain, endpoint string, data interface{}) (*http.Request, error) {
	req, err := NewRequest(method, subdomain, endpoint, data)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("auth_token %s", auth.AccessToken))
	return req, nil
}

func NewRequest(method, subdomain, endpoint string, data interface{}) (*http.Request, error) {
	url := fmt.Sprintf("https://%s.aylanetworks.com/%s", subdomain, endpoint)

	payload, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	dataReader := bytes.NewReader(payload)

	req, err := http.NewRequest(method, url, dataReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func (c *Client) getMiniToken(idToken string) (string, error) {
	url := "https://ayla-sso.eu.owletdata.com/mini/"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", idToken)

	resp, err := c.implementationClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to request mini token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	dict := map[string]string{}
	if err := json.NewDecoder(resp.Body).Decode(&dict); err != nil {
		return "", fmt.Errorf("failed to parse mini token response: %w", err)
	}

	return dict["mini_token"], nil
}

func (c *Client) tokenSignIn(miniToken string) error {
	data := map[string]interface{}{
		"app_id":     OwletAppID,
		"app_secret": OwletAppSecret,
		"provider":   "owl_id",
		"token":      miniToken,
	}

	req, err := NewRequest("POST", UserFieldSubdomain, "api/v1/token_sign_in", data)
	if err != nil {
		return fmt.Errorf("failed to create new request: %w", err)
	}

	resp, err := c.implementationClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to request token sign in: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read token signin response: %w", err)
	}

	auth := &Authentication{}
	err = json.Unmarshal(body, auth)
	if err != nil {
		return fmt.Errorf("failed to parse token signin response.", body, err)
	}

	c.Auth = auth
	c.TokenExpiry = time.Now().Add(time.Duration(auth.ExpiresIn) * time.Second)

	return nil
}

func (c *Client) refreshAuth() error {

	if c.Auth.RefreshToken == "" {
		return fmt.Errorf("refresh token is empty")
	}

	fmt.Println("Refreshing token...")

	url := fmt.Sprintf("https://securetoken.googleapis.com/v1/token?key=%s", ApiKey)

	data := map[string]string{
		"grantType":    "refresh_token",
		"refreshToken": c.Auth.RefreshToken,
	}

	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}

	dataReader := bytes.NewReader(payload)

	req, err := http.NewRequest("POST", url, dataReader)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Android-Package", AndroidPackage)
	req.Header.Set("X-Android-Cert", AndroidCert)

	resp, err := c.implementationClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to request a refresh auth: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	dict := map[string]string{}
	err = json.NewDecoder(resp.Body).Decode(&dict)
	if err != nil {
		return fmt.Errorf("failed to parse refresh auth response: %w", err)
	}

	c.Auth.RefreshToken = dict["refresh_token"]

	miniToken, err := c.getMiniToken(dict["id_token"])
	if err != nil {
		return fmt.Errorf("failed to get mini token: %w", err)
	}

	return c.tokenSignIn(miniToken)
}
