package runtime

import "testing"

func TestOwnershipScopes(t *testing.T) {
	owner := Ownership{InstallationID: "11111111-1111-4111-8111-111111111111", ProjectID: "22222222-2222-4222-8222-222222222222", OperationID: "33333333-3333-4333-8333-333333333333", Scope: "project"}
	if err := owner.Validate(); err != nil {
		t.Fatal(err)
	}
	owner.Scope = "installation"
	if owner.Validate() == nil {
		t.Fatal("installation accepted project identity")
	}
	owner.ProjectID = owner.InstallationID
	if err := owner.Validate(); err != nil {
		t.Fatal(err)
	}
	owner.Scope = "project"
	if owner.Validate() == nil {
		t.Fatal("project accepted installation identity")
	}
	owner.Scope = "installation"
	owner.OperationID = "invalid"
	if owner.Validate() == nil {
		t.Fatal("invalid operation accepted")
	}
}

func TestOwnershipValidatesEntireLedger(t *testing.T) {
	base := Ownership{InstallationID: "11111111-1111-4111-8111-111111111111", ProjectID: "22222222-2222-4222-8222-222222222222", OperationID: "33333333-3333-4333-8333-333333333333", Scope: "project"}
	resource := Resource{Kind: "container", ID: "web", Name: "web", Role: "web", InstallationID: base.InstallationID, ProjectID: base.ProjectID, CreationOperation: base.OperationID}
	cases := map[string]func(*Resource){"kind": func(r *Resource) { r.Kind = "" }, "id": func(r *Resource) { r.ID = "" }, "name": func(r *Resource) { r.Name = "" }, "role": func(r *Resource) { r.Role = "" }, "installation": func(r *Resource) { r.InstallationID = "foreign" }, "project": func(r *Resource) { r.ProjectID = "foreign" }, "operation": func(r *Resource) { r.CreationOperation = "bad" }, "planned-tombstone": func(r *Resource) { r.Planned = true; r.Deleted = true }}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			o := base
			r := resource
			change(&r)
			o.Resources = []Resource{r}
			if o.Validate() == nil {
				t.Fatal("invalid resource accepted")
			}
		})
	}
	base.Resources = []Resource{resource, resource}
	if base.Validate() == nil {
		t.Fatal("duplicate incarnation accepted")
	}
	next := resource
	next.CreationOperation = "44444444-4444-4444-8444-444444444444"
	base.Resources = []Resource{resource, next}
	if base.Validate() == nil {
		t.Fatal("two live incarnations accepted")
	}
	base.Resources[0].Deleted = true
	if err := base.Validate(); err != nil {
		t.Fatal(err)
	}
}
