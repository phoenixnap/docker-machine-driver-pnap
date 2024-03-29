<h1 align="center">
  <br>
  <a href="https://phoenixnap.com/bare-metal-cloud"><img src="https://user-images.githubusercontent.com/78744488/109779287-16da8600-7c06-11eb-81a1-97bf44983d33.png" alt="phoenixnap Bare Metal Cloud" width="300"></a>
  <br>
  Docker Machine Driver Plugin for Bare Metal Cloud
  <br>
</h1>

<p align="center">
Create Docker Machines on Bare Metal Cloud.
</p>

<p align="center">
  <a href="https://phoenixnap.com/bare-metal-cloud">Bare Metal Cloud</a> •
  <a href="https://developers.phoenixnap.com/apis">API</a> •
  <a href="https://developers.phoenixnap.com/">Developers Portal</a> •
  <a href="http://phoenixnap.com/kb">Knowledge Base</a> •
  <a href="https://developers.phoenixnap.com/support">Support</a>
</p>

## Requirements

- [Bare Metal Cloud](https://bmc.phoenixnap.com) account
- [Go](https://golang.org/dl/)

## Creating a Bare Metal Cloud account

1. Go to the [Bare Metal Cloud signup page](https://support.phoenixnap.com/wap-jpost3/bmcSignup).
2. Follow the prompts to set up your account.
3. Use your credentials to [log in to Bare Metal Cloud portal](https://bmc.phoenixnap.com).

:arrow_forward: **Video tutorial:** [How to Create a Bare Metal Cloud Account in Minutes](https://www.youtube.com/watch?v=hPR60XWOSsQ)
<br>

:arrow_forward: **Video tutorial:** [How to Deploy a Bare Metal Server in a Minute](https://www.youtube.com/watch?v=BzBBwLxR80o)

## Available functions

- `NewDriver()`: creates and returns a new instance of the PNAP driver
- `getClient()`: creates the pnap API Client
- `GetCreateFlags()`:  returns the mcnflag.Flag slice representing the flags that can be set, their descriptions and defaults.
- `DriverName()`: returns the name of the driver
- `SetConfigFromFlags()`: configures the driver with the object that was returned by RegisterCreateFlags
- `createSSHKey()`: creates SSH key
- `publicSSHKeyPath()`: SSH key path appended with ".pub"
- `Create()`: creates a host using the driver's config
- `GetSSHHostname()`: returns hostname for use with ssh
- `GetState()`: returns the state that the host is in (running, stopped, etc)
- `Kill()`: force stop host
- `Remove()`: removes host
- `Restart()`: restarts host - this may just call Stop(); Start() if the provider does not have any special restart behavior
- `Start()`: starts host
- `Stop()`: force stop host
- `PreCreateCheck()`: allows for pre-create operations to make sure a driver is ready for creation
- `GetSSHUsername()`: returns username for use with ssh

## API Credentials

Follow these steps to obtain your API authentication credentials.

1. [Log in to the Bare Metal Cloud portal](https://bmc.phoenixnap.com).
2. On the left side menu, click on API Credentials.
3. Click the Create Credentials button.
4. Fill in the Name and Description fields, select the permissions scope and click Create.
5. In the table, click on Actions and select View Credentials from the dropdown to view the Client ID and Client Secret.

**Bare Metal Cloud Quick Start Guide**: [https://developers.phoenixnap.com/quick-start](https://developers.phoenixnap.com/quick-start)

## Bare Metal Cloud community

Become part of the Bare Metal Cloud community to get updates on new features, help us improve the platform, and engage with developers and other users.

- Follow [@phoenixNAP on Twitter](https://twitter.com/phoenixnap)
- Join the [official Slack channel](https://phoenixnap.slack.com)
- Sign up for our [Developers Monthly newsletter](https://phoenixnap.com/developers-monthly-newsletter)

### Resources

- [Product page](https://phoenixnap.com/bare-metal-cloud)
- [Instance pricing](https://phoenixnap.com/bare-metal-cloud/instances)
- [YouTube tutorials](https://www.youtube.com/watch?v=8TLsqgLDMN4&list=PLWcrQnFWd54WwkHM0oPpR1BrAhxlsy1Rc&ab_channel=PhoenixNAPGlobalITServices)
- [Developers Portal](https://developers.phoenixnap.com)
- [Knowledge Base](https://phoenixnap.com/kb)
- [Blog](https:/phoenixnap.com/blog)

### Documentation

- [API documentation](https://developers.phoenixnap.com/apis)

### Contact phoenixNAP

Get in touch with us if you have questions or need help with Bare Metal Cloud.

<p align="left">
  <a href="https://twitter.com/phoenixNAP">Twitter</a> •
  <a href="https://www.facebook.com/phoenixnap">Facebook</a> •
  <a href="https://www.linkedin.com/company/phoenix-nap">LinkedIn</a> •
  <a href="https://www.instagram.com/phoenixnap">Instagram</a> •
  <a href="https://www.youtube.com/user/PhoenixNAPdatacenter">YouTube</a> •
  <a href="https://developers.phoenixnap.com/support">Email</a> 
</p>

<p align="center">
  <br>
  <a href="https://phoenixnap.com/bare-metal-cloud"><img src="https://user-images.githubusercontent.com/78744488/109779474-47222480-7c06-11eb-8ed6-91e28af3a79c.jpg" alt="phoenixnap Bare Metal Cloud"></a>
</p>
