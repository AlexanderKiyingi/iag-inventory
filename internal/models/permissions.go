package models

type PermissionDescriptor struct {
	Name        string
	Description string
}

// PermissionDescriptors is the catalogue registered with iag-authentication at
// boot. Seeded with the baseline read/write permissions; expanded as the
// inventory domain (SKU master, ledger, movements) is implemented.
func PermissionDescriptors() []PermissionDescriptor {
	return []PermissionDescriptor{
		{Name: "inventory.view_overview", Description: "View inventory service overview and status"},
		{Name: "inventory.view_stock", Description: "View on-hand stock levels and SKU attributes"},
		{Name: "inventory.change_stock", Description: "Record stock movements and adjustments"},
	}
}
