package pkg

type GatewayI interface {
	Login() error
	Reboot(dryRun bool) error
}
