package notification

import "github.com/ecletus/admin"

func AdminSetup(Admin *admin.Admin) {
	Admin.NewResource(&Notification{}, &admin.Config{
		Virtual:true,
		Setup: func(res *admin.Resource) {

		},
	})
}
