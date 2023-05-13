The `authy-cli` tool is an alternative client for connecting to Authy.  It is meant for educational purposes only.

# Commands

```
$ authy-cli -help
USAGE
  authy-cli [flags] <subcommand>

SUBCOMMANDS
  register  Register as a device on this account
  sync      Sync your tokens from Authy
  export    export <file>
  token     Get a OTP
  list      List out all known tokens
  unlock    unlock

FLAGS
  -har=false     Always write all HTTP requests to a HAR file, not just on error.
  -password ...  Password to use for the storage files. Can also be set with AUTHY_CLI_STORAGE_PASSWORD environment variable.

$
```

## register

The first step is to register `authy-cli` as a new device.

You should provide your contry code and phone number, with no dashes or spaces, numbers only.

```
$ authy-cli register 1 5558675309
2023/05/12 20:35:43 Checking status of account 1-5558675309
2023/05/12 20:35:43 Account ID is 123456
2023/05/12 20:35:43 Device registration request sent to other devices via push
2023/05/12 20:35:43 Please accept this request.
2023/05/12 20:35:51 Registration approved!
Enter Storage Password:
2023/05/12 20:35:52 Registration complete!
```


## sync

Once registered, sync will pull down all your tokens and apps.

```
$ authy-cli sync
Enter Storage Password:
2023/05/12 20:09:15 Synced 31 tokens
2023/05/12 20:09:15 Synced 6 apps
```

## list

Once synced, list will display all your tokens and apps in a concise format.

```
$ authy-cli list
Enter Storage Password:
[ Authenticator Tokens ]

                Type |                        Name |         ID
---------------------|-----------------------------|-----------
              stripe | Stripe : nobody@example.com | 1111111111
       authenticator |                        Blog | 2222222222
              google |          nobody@exammple.co | 3333333333

[ Authy Apps ]

         Name |                       ID
--------------|-------------------------
   Cloudflare | 000000000000000000000000
 Code Climate | 111111111111111111111111
     SendGrid | 222222222222222222222222
```

## token

Token will take a token or app ID or Name and present the current and next token values.

```
$ authy-cli token 1111111111
Enter Storage Password:
Enter Backup Password:
Current: 699902 (17s)
   Next: 600695
```

## export

Export will dump all of the information `authy-cli` holds as a JSON object onto stdout.

```
$ authy-cli export | jq .
Enter Storage Password:
{
  "Config": {
    "authy_id": 123456,
    "device": {
      "id": 654321,
      "secret_seed": "abcde",
      "api_key": "abcde"
    }
  },
  "Tokens": [
    {
      "account_type": "slack",
      ...
```

## unlock

Unlock will store your backups password on the config object, allowing you to get a token or plain text export without providing the password.

```
$ authy-cli unlock
Enter Storage Password:
Enter Backup Password:
$ authy-cli token 1111111111
Enter Storage Password:
Current: 699902 (17s)
   Next: 600695
```

# Storage

Your Authy credentials and tokens are stored on disc in `$HOME/.authy-cli`.  They are encrypted with a storage password you set with AES.  See `store/store.go` for details.

# Debugging

This tool is rough and ready, but fairly sturdy.  In the event of an error, it will attempt to write all of the API requests it handled into a HAR file.  This can be forced to write in non-error situations by passing the `-har` flag to any command which communicates with the API.

```
$ ./authy-cli sync -har
Enter Storage Password:
2023/05/12 20:40:05 Synced 31 tokens
2023/05/12 20:40:06 Synced 6 apps
2023/05/12 20:40:07 Writing all HTTP requests to 2023-05-12T20:40:01.har
2023/05/12 20:40:07 !!! THIS HAR MAY CONTAIN SENSITIVE INFORMATION !!!
$
```

# Password

There are two passwords involved in this application.  The Authy Backup Password, and the Authy CLI Storage Password.

The Authy Backup Password is the password used to secure your encrypted auth tokens which sync between devices.  This is created and lives external to the authy-cli.  It can be stored using the `unlock` command, or you can enter it whenever you are prompted.

The Authy CLI Storage Password is used to secure the data pulled from Authy at rest on your computer.  You will be prompted to create this password when complete the registration phase.  You can use the `-password` flag, or the `AUTHY_CLI_STORAGE_PASSWORD` environment variable to avoid entering it manually.


