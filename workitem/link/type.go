package link

import (
	"github.com/almighty/almighty-core/app"
	convert "github.com/almighty/almighty-core/convert"
	"github.com/almighty/almighty-core/errors"
	"github.com/almighty/almighty-core/gormsupport"
	"github.com/almighty/almighty-core/rest"

	"github.com/goadesign/goa"
	errs "github.com/pkg/errors"
	satoriuuid "github.com/satori/go.uuid"
)

const (
	TopologyNetwork         = "network"
	TopologyDirectedNetwork = "directed_network"
	TopologyDependency      = "dependency"
	TopologyTree            = "tree"

	// The names of a work item link type are basically the "system.title" field
	// as in work items. The actual linking is done with UUIDs. Hence, the names
	// hare are more human-readable.
	SystemWorkItemLinkTypeBugBlocker     = "Bug blocker"
	SystemWorkItemLinkPlannerItemRelated = "Related planner item"
)

// returns true if the left hand and right hand side string
// pointers either both point to nil or reference the same
// content; otherwise false is returned.
func strPtrIsNilOrContentIsEqual(l, r *string) bool {
	if l == nil && r != nil {
		return false
	}
	if l != nil && r == nil {
		return false
	}
	if l == nil && r == nil {
		return true
	}
	return *l == *r
}

// WorkItemLinkType represents the type of a work item link as it is stored in the db
type WorkItemLinkType struct {
	gormsupport.Lifecycle
	// ID
	ID satoriuuid.UUID `sql:"type:uuid default uuid_generate_v4()" gorm:"primary_key"`
	// Name is the unique name of this work item link type.
	Name string
	// Description is an optional description of the work item link type
	Description *string
	// Version for optimistic concurrency control
	Version  int
	Topology string // Valid values: network, directed_network, dependency, tree

	SourceTypeID satoriuuid.UUID `sql:"type:uuid"`
	TargetTypeID satoriuuid.UUID `sql:"type:uuid"`

	ForwardName string
	ReverseName string

	LinkCategoryID satoriuuid.UUID `sql:"type:uuid"`

	// Reference to one Space
	SpaceID satoriuuid.UUID `sql:"type:uuid"`
}

// Ensure Fields implements the Equaler interface
var _ convert.Equaler = WorkItemLinkType{}
var _ convert.Equaler = (*WorkItemLinkType)(nil)

// Equal returns true if two WorkItemLinkType objects are equal; otherwise false is returned.
func (t WorkItemLinkType) Equal(u convert.Equaler) bool {
	other, ok := u.(WorkItemLinkType)
	if !ok {
		return false
	}
	if !t.Lifecycle.Equal(other.Lifecycle) {
		return false
	}
	if !satoriuuid.Equal(t.ID, other.ID) {
		return false
	}
	if t.Name != other.Name {
		return false
	}
	if t.Version != other.Version {
		return false
	}
	if !strPtrIsNilOrContentIsEqual(t.Description, other.Description) {
		return false
	}
	if t.Topology != other.Topology {
		return false
	}
	if !satoriuuid.Equal(t.SourceTypeID, other.SourceTypeID) {
		return false
	}
	if !satoriuuid.Equal(t.TargetTypeID, other.TargetTypeID) {
		return false
	}
	if t.ForwardName != other.ForwardName {
		return false
	}
	if t.ReverseName != other.ReverseName {
		return false
	}
	if !satoriuuid.Equal(t.LinkCategoryID, other.LinkCategoryID) {
		return false
	}
	if !satoriuuid.Equal(t.SpaceID, other.SpaceID) {
		return false
	}
	return true
}

// CheckValidForCreation returns an error if the work item link type
// cannot be used for the creation of a new work item link type.
func (t *WorkItemLinkType) CheckValidForCreation() error {
	if t.Name == "" {
		return errors.NewBadParameterError("name", t.Name)
	}
	if satoriuuid.Equal(t.SourceTypeID, satoriuuid.Nil) {
		return errors.NewBadParameterError("source_type_name", t.SourceTypeID)
	}
	if satoriuuid.Equal(t.TargetTypeID, satoriuuid.Nil) {
		return errors.NewBadParameterError("target_type_name", t.TargetTypeID)
	}
	if t.ForwardName == "" {
		return errors.NewBadParameterError("forward_name", t.ForwardName)
	}
	if t.ReverseName == "" {
		return errors.NewBadParameterError("reverse_name", t.ReverseName)
	}
	if err := CheckValidTopology(t.Topology); err != nil {
		return errs.WithStack(err)
	}
	if t.LinkCategoryID == satoriuuid.Nil {
		return errors.NewBadParameterError("link_category_id", t.LinkCategoryID)
	}
	if t.SpaceID == satoriuuid.Nil {
		return errors.NewBadParameterError("space_id", t.SpaceID)
	}
	return nil
}

// TableName implements gorm.tabler
func (t WorkItemLinkType) TableName() string {
	return "work_item_link_types"
}

// CheckValidTopology returns nil if the given topology is valid;
// otherwise a BadParameterError is returned.
func CheckValidTopology(t string) error {
	if t != TopologyNetwork && t != TopologyDirectedNetwork && t != TopologyDependency && t != TopologyTree {
		return errors.NewBadParameterError("topolgy", t).Expected(TopologyNetwork + "|" + TopologyDirectedNetwork + "|" + TopologyDependency + "|" + TopologyTree)
	}
	return nil
}

// ConvertLinkTypeFromModel converts a work item link type from model to REST representation
func ConvertLinkTypeFromModel(request *goa.RequestData, t WorkItemLinkType) app.WorkItemLinkTypeSingle {
	spaceType := "spaces"
	spaceSelfURL := rest.AbsoluteURL(request, app.SpaceHref(t.SpaceID.String()))

	var converted = app.WorkItemLinkTypeSingle{
		Data: &app.WorkItemLinkTypeData{
			Type: EndpointWorkItemLinkTypes,
			ID:   &t.ID,
			Attributes: &app.WorkItemLinkTypeAttributes{
				Name:        &t.Name,
				Description: t.Description,
				Version:     &t.Version,
				ForwardName: &t.ForwardName,
				ReverseName: &t.ReverseName,
				Topology:    &t.Topology,
			},
			Relationships: &app.WorkItemLinkTypeRelationships{
				LinkCategory: &app.RelationWorkItemLinkCategory{
					Data: &app.RelationWorkItemLinkCategoryData{
						Type: EndpointWorkItemLinkCategories,
						ID:   t.LinkCategoryID,
					},
				},
				SourceType: &app.RelationWorkItemType{
					Data: &app.RelationWorkItemTypeData{
						Type: EndpointWorkItemTypes,
						ID:   t.SourceTypeID,
					},
				},
				TargetType: &app.RelationWorkItemType{
					Data: &app.RelationWorkItemTypeData{
						Type: EndpointWorkItemTypes,
						ID:   t.TargetTypeID,
					},
				},
				Space: &app.RelationSpaces{
					Data: &app.RelationSpacesData{
						Type: &spaceType,
						ID:   &t.SpaceID,
					},
					Links: &app.GenericLinks{
						Self: &spaceSelfURL,
					},
				},
			},
		},
	}
	return converted
}

// ConvertLinkTypeToModel converts the incoming app representation of a work item link type to the model layout.
// Values are only overwrriten if they are set in "in", otherwise the values in "out" remain.
func ConvertLinkTypeToModel(in app.WorkItemLinkTypeSingle, out *WorkItemLinkType) error {
	if in.Data == nil {
		return errors.NewBadParameterError("data", nil).Expected("not <nil>")
	}
	if in.Data.Attributes == nil {
		return errors.NewBadParameterError("data.attributes", nil).Expected("not <nil>")
	}
	if in.Data.Relationships == nil {
		return errors.NewBadParameterError("data.relationships", nil).Expected("not <nil>")
	}

	attrs := in.Data.Attributes
	rel := in.Data.Relationships

	if in.Data.ID != nil {
		out.ID = *in.Data.ID
	}

	if attrs != nil {
		// If the name is not nil, it MUST NOT be empty
		if attrs.Name != nil {
			if *attrs.Name == "" {
				return errors.NewBadParameterError("data.attributes.name", *attrs.Name)
			}
			out.Name = *attrs.Name
		}

		if attrs.Description != nil {
			out.Description = attrs.Description
		}

		if attrs.Version != nil {
			out.Version = *attrs.Version
		}

		// If the forwardName is not nil, it MUST NOT be empty
		if attrs.ForwardName != nil {
			if *attrs.ForwardName == "" {
				return errors.NewBadParameterError("data.attributes.forward_name", *attrs.ForwardName)
			}
			out.ForwardName = *attrs.ForwardName
		}

		// If the ReverseName is not nil, it MUST NOT be empty
		if attrs.ReverseName != nil {
			if *attrs.ReverseName == "" {
				return errors.NewBadParameterError("data.attributes.reverse_name", *attrs.ReverseName)
			}
			out.ReverseName = *attrs.ReverseName
		}

		if attrs.Topology != nil {
			if err := CheckValidTopology(*attrs.Topology); err != nil {
				return errs.WithStack(err)
			}
			out.Topology = *attrs.Topology
		}
	}

	if rel != nil && rel.LinkCategory != nil && rel.LinkCategory.Data != nil {
		out.LinkCategoryID = rel.LinkCategory.Data.ID
	}
	if rel != nil && rel.SourceType != nil && rel.SourceType.Data != nil {
		out.SourceTypeID = rel.SourceType.Data.ID
	}
	if rel != nil && rel.TargetType != nil && rel.TargetType.Data != nil {
		out.TargetTypeID = rel.TargetType.Data.ID
	}
	if rel != nil && rel.Space != nil && rel.Space.Data != nil {
		out.SpaceID = *rel.Space.Data.ID
	}

	return nil
}
