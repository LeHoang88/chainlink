package resolver

import (
	"database/sql"
	"errors"
	"strconv"
	"strings"

	"github.com/graph-gophers/graphql-go"

	"github.com/smartcontractkit/chainlink/core/services/feeds"
)

type JobType string

func ToJobType(s string) (JobType, error) {
	switch s {
	case "fluxmonitor":
		return JobTypeFluxMonitor, nil
	case "ocr":
		return JobTypeOCR, nil
	default:
		return JobType(""), errors.New("invalid job type")
	}
}

// FromJobType converts a JoyType into a string which is used to create/update
// a feeds manager.
//
// FluxMonitor is a special case because the FeedsManager expects a 'fluxmonitor'
// string rather than just the lowercased value.
//
// Note: We should change the flux monitor service methods to accept an enum for
// the job type to add more consistency.
func FromJobTypeInput(jt JobType) string {
	switch jt {
	// Handle the special case of flux monitor string
	case JobTypeFluxMonitor:
		return "fluxmonitor"
	default:
		return strings.ToLower(string(jt))
	}
}

const (
	JobTypeFluxMonitor JobType = "FLUX_MONITOR"
	JobTypeOCR         JobType = "OCR"
)

// FeedsManagerResolver resolves the FeedsManager type.
type FeedsManagerResolver struct {
	mgr feeds.FeedsManager
}

func NewFeedsManager(mgr feeds.FeedsManager) *FeedsManagerResolver {
	return &FeedsManagerResolver{mgr: mgr}
}

func NewFeedsManagers(mgrs []feeds.FeedsManager) []*FeedsManagerResolver {
	resolvers := []*FeedsManagerResolver{}
	for _, mgr := range mgrs {
		resolvers = append(resolvers, NewFeedsManager(mgr))
	}

	return resolvers
}

// ID resolves the feed managers's unique identifier.
func (r *FeedsManagerResolver) ID() graphql.ID {
	return graphql.ID(strconv.FormatInt(r.mgr.ID, 10))
}

// Name resolves the feed managers's name field.
func (r *FeedsManagerResolver) Name() string {
	return r.mgr.Name
}

// URI resolves the feed managers's uri field.
func (r *FeedsManagerResolver) URI() string {
	return r.mgr.URI
}

// PublicKey resolves the feed managers's public key field.
func (r *FeedsManagerResolver) PublicKey() string {
	return r.mgr.PublicKey.String()
}

// JobTypes resolves the feed managers's jobTypes field.
func (r *FeedsManagerResolver) JobTypes() []JobType {
	jts := []JobType{}

	for _, s := range r.mgr.JobTypes {
		if jt, err := ToJobType(s); err == nil {
			jts = append(jts, jt)
		}
	}

	return jts
}

// IsBootstrapPeer resolves the feed managers's isBootstrapPeer field.
func (r *FeedsManagerResolver) IsBootstrapPeer() bool {
	return r.mgr.IsOCRBootstrapPeer
}

// IsConnectionActive resolves the feed managers's isConnectionActive field.
func (r *FeedsManagerResolver) IsConnectionActive() bool {
	return r.mgr.IsConnectionActive
}

// BootstrapPeer resolves the feed managers's isConnectionActive field.
func (r *FeedsManagerResolver) BootstrapPeerMultiaddr() *string {
	return r.mgr.OCRBootstrapPeerMultiaddr.Ptr()
}

// CreatedAt resolves the chains's created at field.
func (r *FeedsManagerResolver) CreatedAt() graphql.Time {
	return graphql.Time{Time: r.mgr.CreatedAt}
}

// -- FeedsManager Query --

type FeedsManagerPayloadResolver struct {
	mgr *feeds.FeedsManager
}

func NewFeedsManagerPayload(mgr *feeds.FeedsManager) *FeedsManagerPayloadResolver {
	return &FeedsManagerPayloadResolver{mgr: mgr}
}

// ToFeedsManager implements the FeedsManager union type of the payload
func (r *FeedsManagerPayloadResolver) ToFeedsManager() (*FeedsManagerResolver, bool) {
	if r.mgr != nil {
		return NewFeedsManager(*r.mgr), true
	}

	return nil, false
}

// ToNotFoundError implements the NotFoundError union type of the payload
func (r *FeedsManagerPayloadResolver) ToNotFoundError() (*NotFoundErrorResolver, bool) {
	if r.mgr == nil {
		return NewNotFoundError("feeds manager not found"), true
	}

	return nil, false
}

// -- FeedsManagers Query --

// FeedsManagersPayloadResolver resolves a list of feeds managers
type FeedsManagersPayloadResolver struct {
	feedsManagers []feeds.FeedsManager
}

func NewFeedsManagersPayload(feedsManagers []feeds.FeedsManager) *FeedsManagersPayloadResolver {
	return &FeedsManagersPayloadResolver{
		feedsManagers: feedsManagers,
	}
}

// Results returns the feeds managers.
func (r *FeedsManagersPayloadResolver) Results() []*FeedsManagerResolver {
	return NewFeedsManagers(r.feedsManagers)
}

// -- CreateFeedsManager Mutation --

// CreateFeedsManagerPayloadResolver
type CreateFeedsManagerPayloadResolver struct {
	mgr *feeds.FeedsManager
	err error
	// inputErrors maps an input path to a string
	inputErrs map[string]string
}

func NewCreateFeedsManagerPayload(mgr *feeds.FeedsManager, err error, inputErrs map[string]string) *CreateFeedsManagerPayloadResolver {
	return &CreateFeedsManagerPayloadResolver{
		mgr:       mgr,
		err:       err,
		inputErrs: inputErrs,
	}
}

func (r *CreateFeedsManagerPayloadResolver) ToCreateFeedsManagerSuccess() (*CreateFeedsManagerSuccessResolver, bool) {
	if r.mgr != nil {
		return NewCreateFeedsManagerSuccessResolver(*r.mgr), true
	}

	return nil, false
}

func (r *CreateFeedsManagerPayloadResolver) ToSingleFeedsManagerError() (*SingleFeedsManagerErrorResolver, bool) {
	if r.err != nil && errors.Is(r.err, feeds.ErrSingleFeedsManager) {
		return NewSingleFeedsManagerError(r.err.Error()), true
	}

	return nil, false
}

func (r *CreateFeedsManagerPayloadResolver) ToNotFoundError() (*NotFoundErrorResolver, bool) {
	if r.err != nil && errors.Is(r.err, sql.ErrNoRows) {
		return NewNotFoundError("feeds manager not found"), true
	}

	return nil, false
}

func (r *CreateFeedsManagerPayloadResolver) ToInputErrors() (*InputErrorsResolver, bool) {
	if r.inputErrs != nil {
		errs := []*InputErrorResolver{}

		for path, message := range r.inputErrs {
			errs = append(errs, NewInputError(path, message))
		}

		return NewInputErrors(errs), true
	}

	return nil, false
}

type CreateFeedsManagerSuccessResolver struct {
	mgr feeds.FeedsManager
}

func NewCreateFeedsManagerSuccessResolver(mgr feeds.FeedsManager) *CreateFeedsManagerSuccessResolver {
	return &CreateFeedsManagerSuccessResolver{
		mgr: mgr,
	}
}

func (r *CreateFeedsManagerSuccessResolver) FeedsManager() *FeedsManagerResolver {
	return NewFeedsManager(r.mgr)
}

type SingleFeedsManagerErrorResolver struct {
	message string
}

func NewSingleFeedsManagerError(message string) *SingleFeedsManagerErrorResolver {
	return &SingleFeedsManagerErrorResolver{
		message: message,
	}
}

func (r *SingleFeedsManagerErrorResolver) Message() string {
	return r.message
}

func (r *SingleFeedsManagerErrorResolver) Code() ErrorCode {
	return ErrorCodeUnprocessable
}

// -- UpdateFeedsManager Mutation --

// UpdateFeedsManagerPayloadResolver
type UpdateFeedsManagerPayloadResolver struct {
	mgr       *feeds.FeedsManager
	err       error
	inputErrs map[string]string
}

func NewUpdateFeedsManagerPayload(mgr *feeds.FeedsManager, err error, inputErrs map[string]string) *UpdateFeedsManagerPayloadResolver {
	return &UpdateFeedsManagerPayloadResolver{
		mgr:       mgr,
		err:       err,
		inputErrs: inputErrs,
	}
}

func (r *UpdateFeedsManagerPayloadResolver) ToUpdateFeedsManagerSuccess() (*UpdateFeedsManagerSuccessResolver, bool) {
	if r.mgr != nil {
		return NewUpdateFeedsManagerSuccessResolver(*r.mgr), true
	}

	return nil, false
}

func (r *UpdateFeedsManagerPayloadResolver) ToNotFoundError() (*NotFoundErrorResolver, bool) {
	if r.err != nil && errors.Is(r.err, sql.ErrNoRows) {
		return NewNotFoundError("feeds manager not found"), true
	}

	return nil, false
}

func (r *UpdateFeedsManagerPayloadResolver) ToInputErrors() (*InputErrorsResolver, bool) {
	if r.inputErrs != nil {
		errs := []*InputErrorResolver{}

		for path, message := range r.inputErrs {
			errs = append(errs, NewInputError(path, message))
		}

		return NewInputErrors(errs), true
	}

	return nil, false
}

type UpdateFeedsManagerSuccessResolver struct {
	mgr feeds.FeedsManager
}

func NewUpdateFeedsManagerSuccessResolver(mgr feeds.FeedsManager) *UpdateFeedsManagerSuccessResolver {
	return &UpdateFeedsManagerSuccessResolver{
		mgr: mgr,
	}
}

func (r *UpdateFeedsManagerSuccessResolver) FeedsManager() *FeedsManagerResolver {
	return NewFeedsManager(r.mgr)
}
