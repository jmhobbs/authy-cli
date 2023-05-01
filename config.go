package main

type Config struct {
	AuthyId int    `json:"authy_id"`
	Device  Device `json:"device"`
}

type Device struct {
	Id         int    `json:"id"`
	SecretSeed string `json:"secret_seed"`
	ApiKey     string `json:"api_key"`
}

type Account struct {
	Email                string `json:"email"`
	CellPhone            string `json:"cellphone"`
	CountryCode          int    `json:"country_code"`
	MultiDeviceEnabled   bool   `json:"multidevice_enabled"`
	MultiDevicesEnabled  bool   `json:"multidevices_enabled"`
	PrimaryEmailVerified bool   `json:"primary_email_verified"`
}

type SoftToken struct {
	Id           string `json:"unique_id"`
	Type         string `json:"account_type"`
	Name         string `json:"name"`
	OriginalName string `json:"original_name"`
	Digits       int    `json:"digits"`
	Seed         string `json:"encrypted_seed"`
}
