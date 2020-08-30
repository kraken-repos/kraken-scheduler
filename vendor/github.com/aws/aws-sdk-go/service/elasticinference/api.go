// Code generated by private/model/cli/gen-api/main.go. DO NOT EDIT.

package elasticinference

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/private/protocol"
	"github.com/aws/aws-sdk-go/private/protocol/restjson"
)

const opListTagsForResource = "ListTagsForResource"

// ListTagsForResourceRequest generates a "aws/request.Request" representing the
// client's request for the ListTagsForResource operation. The "output" return
// value will be populated with the request's response once the request completes
// successfully.
//
// Use "Send" method on the returned Request to send the API call to the service.
// the "output" return value is not valid until after Send returns without error.
//
// See ListTagsForResource for more information on using the ListTagsForResource
// API call, and error handling.
//
// This method is useful when you want to inject custom logic or configuration
// into the SDK's request lifecycle. Such as custom headers, or retry logic.
//
//
//    // Example sending a request using the ListTagsForResourceRequest method.
//    req, resp := client.ListTagsForResourceRequest(params)
//
//    err := req.Send()
//    if err == nil { // resp is now filled
//        fmt.Println(resp)
//    }
//
// See also, https://docs.aws.amazon.com/goto/WebAPI/elastic-inference-2017-07-25/ListTagsForResource
func (c *ElasticInference) ListTagsForResourceRequest(input *ListTagsForResourceInput) (req *request.Request, output *ListTagsForResourceOutput) {
	op := &request.Operation{
		Name:       opListTagsForResource,
		HTTPMethod: "GET",
		HTTPPath:   "/tags/{resourceArn}",
	}

	if input == nil {
		input = &ListTagsForResourceInput{}
	}

	output = &ListTagsForResourceOutput{}
	req = c.newRequest(op, input, output)
	return
}

// ListTagsForResource API operation for Amazon Elastic  Inference.
//
// Returns all tags of an Elastic Inference Accelerator.
//
// Returns awserr.Error for service API and SDK errors. Use runtime type assertions
// with awserr.Error's Code and Message methods to get detailed information about
// the error.
//
// See the AWS API reference guide for Amazon Elastic  Inference's
// API operation ListTagsForResource for usage and error information.
//
// Returned Error Types:
//   * BadRequestException
//   Raised when a malformed input has been provided to the API.
//
//   * ResourceNotFoundException
//   Raised when the requested resource cannot be found.
//
//   * InternalServerException
//   Raised when an unexpected error occurred during request processing.
//
// See also, https://docs.aws.amazon.com/goto/WebAPI/elastic-inference-2017-07-25/ListTagsForResource
func (c *ElasticInference) ListTagsForResource(input *ListTagsForResourceInput) (*ListTagsForResourceOutput, error) {
	req, out := c.ListTagsForResourceRequest(input)
	return out, req.Send()
}

// ListTagsForResourceWithContext is the same as ListTagsForResource with the addition of
// the ability to pass a context and additional request options.
//
// See ListTagsForResource for details on how to use this API operation.
//
// The context must be non-nil and will be used for request cancellation. If
// the context is nil a panic will occur. In the future the SDK may create
// sub-contexts for http.Requests. See https://golang.org/pkg/context/
// for more information on using Contexts.
func (c *ElasticInference) ListTagsForResourceWithContext(ctx aws.Context, input *ListTagsForResourceInput, opts ...request.Option) (*ListTagsForResourceOutput, error) {
	req, out := c.ListTagsForResourceRequest(input)
	req.SetContext(ctx)
	req.ApplyOptions(opts...)
	return out, req.Send()
}

const opTagResource = "TagResource"

// TagResourceRequest generates a "aws/request.Request" representing the
// client's request for the TagResource operation. The "output" return
// value will be populated with the request's response once the request completes
// successfully.
//
// Use "Send" method on the returned Request to send the API call to the service.
// the "output" return value is not valid until after Send returns without error.
//
// See TagResource for more information on using the TagResource
// API call, and error handling.
//
// This method is useful when you want to inject custom logic or configuration
// into the SDK's request lifecycle. Such as custom headers, or retry logic.
//
//
//    // Example sending a request using the TagResourceRequest method.
//    req, resp := client.TagResourceRequest(params)
//
//    err := req.Send()
//    if err == nil { // resp is now filled
//        fmt.Println(resp)
//    }
//
// See also, https://docs.aws.amazon.com/goto/WebAPI/elastic-inference-2017-07-25/TagResource
func (c *ElasticInference) TagResourceRequest(input *TagResourceInput) (req *request.Request, output *TagResourceOutput) {
	op := &request.Operation{
		Name:       opTagResource,
		HTTPMethod: "POST",
		HTTPPath:   "/tags/{resourceArn}",
	}

	if input == nil {
		input = &TagResourceInput{}
	}

	output = &TagResourceOutput{}
	req = c.newRequest(op, input, output)
	req.Handlers.Unmarshal.Swap(restjson.UnmarshalHandler.Name, protocol.UnmarshalDiscardBodyHandler)
	return
}

// TagResource API operation for Amazon Elastic  Inference.
//
// Adds the specified tag(s) to an Elastic Inference Accelerator.
//
// Returns awserr.Error for service API and SDK errors. Use runtime type assertions
// with awserr.Error's Code and Message methods to get detailed information about
// the error.
//
// See the AWS API reference guide for Amazon Elastic  Inference's
// API operation TagResource for usage and error information.
//
// Returned Error Types:
//   * BadRequestException
//   Raised when a malformed input has been provided to the API.
//
//   * ResourceNotFoundException
//   Raised when the requested resource cannot be found.
//
//   * InternalServerException
//   Raised when an unexpected error occurred during request processing.
//
// See also, https://docs.aws.amazon.com/goto/WebAPI/elastic-inference-2017-07-25/TagResource
func (c *ElasticInference) TagResource(input *TagResourceInput) (*TagResourceOutput, error) {
	req, out := c.TagResourceRequest(input)
	return out, req.Send()
}

// TagResourceWithContext is the same as TagResource with the addition of
// the ability to pass a context and additional request options.
//
// See TagResource for details on how to use this API operation.
//
// The context must be non-nil and will be used for request cancellation. If
// the context is nil a panic will occur. In the future the SDK may create
// sub-contexts for http.Requests. See https://golang.org/pkg/context/
// for more information on using Contexts.
func (c *ElasticInference) TagResourceWithContext(ctx aws.Context, input *TagResourceInput, opts ...request.Option) (*TagResourceOutput, error) {
	req, out := c.TagResourceRequest(input)
	req.SetContext(ctx)
	req.ApplyOptions(opts...)
	return out, req.Send()
}

const opUntagResource = "UntagResource"

// UntagResourceRequest generates a "aws/request.Request" representing the
// client's request for the UntagResource operation. The "output" return
// value will be populated with the request's response once the request completes
// successfully.
//
// Use "Send" method on the returned Request to send the API call to the service.
// the "output" return value is not valid until after Send returns without error.
//
// See UntagResource for more information on using the UntagResource
// API call, and error handling.
//
// This method is useful when you want to inject custom logic or configuration
// into the SDK's request lifecycle. Such as custom headers, or retry logic.
//
//
//    // Example sending a request using the UntagResourceRequest method.
//    req, resp := client.UntagResourceRequest(params)
//
//    err := req.Send()
//    if err == nil { // resp is now filled
//        fmt.Println(resp)
//    }
//
// See also, https://docs.aws.amazon.com/goto/WebAPI/elastic-inference-2017-07-25/UntagResource
func (c *ElasticInference) UntagResourceRequest(input *UntagResourceInput) (req *request.Request, output *UntagResourceOutput) {
	op := &request.Operation{
		Name:       opUntagResource,
		HTTPMethod: "DELETE",
		HTTPPath:   "/tags/{resourceArn}",
	}

	if input == nil {
		input = &UntagResourceInput{}
	}

	output = &UntagResourceOutput{}
	req = c.newRequest(op, input, output)
	req.Handlers.Unmarshal.Swap(restjson.UnmarshalHandler.Name, protocol.UnmarshalDiscardBodyHandler)
	return
}

// UntagResource API operation for Amazon Elastic  Inference.
//
// Removes the specified tag(s) from an Elastic Inference Accelerator.
//
// Returns awserr.Error for service API and SDK errors. Use runtime type assertions
// with awserr.Error's Code and Message methods to get detailed information about
// the error.
//
// See the AWS API reference guide for Amazon Elastic  Inference's
// API operation UntagResource for usage and error information.
//
// Returned Error Types:
//   * BadRequestException
//   Raised when a malformed input has been provided to the API.
//
//   * ResourceNotFoundException
//   Raised when the requested resource cannot be found.
//
//   * InternalServerException
//   Raised when an unexpected error occurred during request processing.
//
// See also, https://docs.aws.amazon.com/goto/WebAPI/elastic-inference-2017-07-25/UntagResource
func (c *ElasticInference) UntagResource(input *UntagResourceInput) (*UntagResourceOutput, error) {
	req, out := c.UntagResourceRequest(input)
	return out, req.Send()
}

// UntagResourceWithContext is the same as UntagResource with the addition of
// the ability to pass a context and additional request options.
//
// See UntagResource for details on how to use this API operation.
//
// The context must be non-nil and will be used for request cancellation. If
// the context is nil a panic will occur. In the future the SDK may create
// sub-contexts for http.Requests. See https://golang.org/pkg/context/
// for more information on using Contexts.
func (c *ElasticInference) UntagResourceWithContext(ctx aws.Context, input *UntagResourceInput, opts ...request.Option) (*UntagResourceOutput, error) {
	req, out := c.UntagResourceRequest(input)
	req.SetContext(ctx)
	req.ApplyOptions(opts...)
	return out, req.Send()
}

// Raised when a malformed input has been provided to the API.
type BadRequestException struct {
	_            struct{}                  `type:"structure"`
	RespMetadata protocol.ResponseMetadata `json:"-" xml:"-"`

	Message_ *string `locationName:"message" type:"string"`
}

// String returns the string representation
func (s BadRequestException) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s BadRequestException) GoString() string {
	return s.String()
}

func newErrorBadRequestException(v protocol.ResponseMetadata) error {
	return &BadRequestException{
		RespMetadata: v,
	}
}

// Code returns the exception type name.
func (s BadRequestException) Code() string {
	return "BadRequestException"
}

// Message returns the exception's message.
func (s BadRequestException) Message() string {
	if s.Message_ != nil {
		return *s.Message_
	}
	return ""
}

// OrigErr always returns nil, satisfies awserr.Error interface.
func (s BadRequestException) OrigErr() error {
	return nil
}

func (s BadRequestException) Error() string {
	return fmt.Sprintf("%s: %s", s.Code(), s.Message())
}

// Status code returns the HTTP status code for the request's response error.
func (s BadRequestException) StatusCode() int {
	return s.RespMetadata.StatusCode
}

// RequestID returns the service's response RequestID for request.
func (s BadRequestException) RequestID() string {
	return s.RespMetadata.RequestID
}

// Raised when an unexpected error occurred during request processing.
type InternalServerException struct {
	_            struct{}                  `type:"structure"`
	RespMetadata protocol.ResponseMetadata `json:"-" xml:"-"`

	Message_ *string `locationName:"message" type:"string"`
}

// String returns the string representation
func (s InternalServerException) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s InternalServerException) GoString() string {
	return s.String()
}

func newErrorInternalServerException(v protocol.ResponseMetadata) error {
	return &InternalServerException{
		RespMetadata: v,
	}
}

// Code returns the exception type name.
func (s InternalServerException) Code() string {
	return "InternalServerException"
}

// Message returns the exception's message.
func (s InternalServerException) Message() string {
	if s.Message_ != nil {
		return *s.Message_
	}
	return ""
}

// OrigErr always returns nil, satisfies awserr.Error interface.
func (s InternalServerException) OrigErr() error {
	return nil
}

func (s InternalServerException) Error() string {
	return fmt.Sprintf("%s: %s", s.Code(), s.Message())
}

// Status code returns the HTTP status code for the request's response error.
func (s InternalServerException) StatusCode() int {
	return s.RespMetadata.StatusCode
}

// RequestID returns the service's response RequestID for request.
func (s InternalServerException) RequestID() string {
	return s.RespMetadata.RequestID
}

type ListTagsForResourceInput struct {
	_ struct{} `type:"structure"`

	// The ARN of the Elastic Inference Accelerator to list the tags for.
	//
	// ResourceArn is a required field
	ResourceArn *string `location:"uri" locationName:"resourceArn" min:"1" type:"string" required:"true"`
}

// String returns the string representation
func (s ListTagsForResourceInput) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s ListTagsForResourceInput) GoString() string {
	return s.String()
}

// Validate inspects the fields of the type to determine if they are valid.
func (s *ListTagsForResourceInput) Validate() error {
	invalidParams := request.ErrInvalidParams{Context: "ListTagsForResourceInput"}
	if s.ResourceArn == nil {
		invalidParams.Add(request.NewErrParamRequired("ResourceArn"))
	}
	if s.ResourceArn != nil && len(*s.ResourceArn) < 1 {
		invalidParams.Add(request.NewErrParamMinLen("ResourceArn", 1))
	}

	if invalidParams.Len() > 0 {
		return invalidParams
	}
	return nil
}

// SetResourceArn sets the ResourceArn field's value.
func (s *ListTagsForResourceInput) SetResourceArn(v string) *ListTagsForResourceInput {
	s.ResourceArn = &v
	return s
}

type ListTagsForResourceOutput struct {
	_ struct{} `type:"structure"`

	// The tags of the Elastic Inference Accelerator.
	Tags map[string]*string `locationName:"tags" min:"1" type:"map"`
}

// String returns the string representation
func (s ListTagsForResourceOutput) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s ListTagsForResourceOutput) GoString() string {
	return s.String()
}

// SetTags sets the Tags field's value.
func (s *ListTagsForResourceOutput) SetTags(v map[string]*string) *ListTagsForResourceOutput {
	s.Tags = v
	return s
}

// Raised when the requested resource cannot be found.
type ResourceNotFoundException struct {
	_            struct{}                  `type:"structure"`
	RespMetadata protocol.ResponseMetadata `json:"-" xml:"-"`

	Message_ *string `locationName:"message" type:"string"`
}

// String returns the string representation
func (s ResourceNotFoundException) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s ResourceNotFoundException) GoString() string {
	return s.String()
}

func newErrorResourceNotFoundException(v protocol.ResponseMetadata) error {
	return &ResourceNotFoundException{
		RespMetadata: v,
	}
}

// Code returns the exception type name.
func (s ResourceNotFoundException) Code() string {
	return "ResourceNotFoundException"
}

// Message returns the exception's message.
func (s ResourceNotFoundException) Message() string {
	if s.Message_ != nil {
		return *s.Message_
	}
	return ""
}

// OrigErr always returns nil, satisfies awserr.Error interface.
func (s ResourceNotFoundException) OrigErr() error {
	return nil
}

func (s ResourceNotFoundException) Error() string {
	return fmt.Sprintf("%s: %s", s.Code(), s.Message())
}

// Status code returns the HTTP status code for the request's response error.
func (s ResourceNotFoundException) StatusCode() int {
	return s.RespMetadata.StatusCode
}

// RequestID returns the service's response RequestID for request.
func (s ResourceNotFoundException) RequestID() string {
	return s.RespMetadata.RequestID
}

type TagResourceInput struct {
	_ struct{} `type:"structure"`

	// The ARN of the Elastic Inference Accelerator to tag.
	//
	// ResourceArn is a required field
	ResourceArn *string `location:"uri" locationName:"resourceArn" min:"1" type:"string" required:"true"`

	// The tags to add to the Elastic Inference Accelerator.
	//
	// Tags is a required field
	Tags map[string]*string `locationName:"tags" min:"1" type:"map" required:"true"`
}

// String returns the string representation
func (s TagResourceInput) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s TagResourceInput) GoString() string {
	return s.String()
}

// Validate inspects the fields of the type to determine if they are valid.
func (s *TagResourceInput) Validate() error {
	invalidParams := request.ErrInvalidParams{Context: "TagResourceInput"}
	if s.ResourceArn == nil {
		invalidParams.Add(request.NewErrParamRequired("ResourceArn"))
	}
	if s.ResourceArn != nil && len(*s.ResourceArn) < 1 {
		invalidParams.Add(request.NewErrParamMinLen("ResourceArn", 1))
	}
	if s.Tags == nil {
		invalidParams.Add(request.NewErrParamRequired("Tags"))
	}
	if s.Tags != nil && len(s.Tags) < 1 {
		invalidParams.Add(request.NewErrParamMinLen("Tags", 1))
	}

	if invalidParams.Len() > 0 {
		return invalidParams
	}
	return nil
}

// SetResourceArn sets the ResourceArn field's value.
func (s *TagResourceInput) SetResourceArn(v string) *TagResourceInput {
	s.ResourceArn = &v
	return s
}

// SetTags sets the Tags field's value.
func (s *TagResourceInput) SetTags(v map[string]*string) *TagResourceInput {
	s.Tags = v
	return s
}

type TagResourceOutput struct {
	_ struct{} `type:"structure"`
}

// String returns the string representation
func (s TagResourceOutput) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s TagResourceOutput) GoString() string {
	return s.String()
}

type UntagResourceInput struct {
	_ struct{} `type:"structure"`

	// The ARN of the Elastic Inference Accelerator to untag.
	//
	// ResourceArn is a required field
	ResourceArn *string `location:"uri" locationName:"resourceArn" min:"1" type:"string" required:"true"`

	// The list of tags to remove from the Elastic Inference Accelerator.
	//
	// TagKeys is a required field
	TagKeys []*string `location:"querystring" locationName:"tagKeys" min:"1" type:"list" required:"true"`
}

// String returns the string representation
func (s UntagResourceInput) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s UntagResourceInput) GoString() string {
	return s.String()
}

// Validate inspects the fields of the type to determine if they are valid.
func (s *UntagResourceInput) Validate() error {
	invalidParams := request.ErrInvalidParams{Context: "UntagResourceInput"}
	if s.ResourceArn == nil {
		invalidParams.Add(request.NewErrParamRequired("ResourceArn"))
	}
	if s.ResourceArn != nil && len(*s.ResourceArn) < 1 {
		invalidParams.Add(request.NewErrParamMinLen("ResourceArn", 1))
	}
	if s.TagKeys == nil {
		invalidParams.Add(request.NewErrParamRequired("TagKeys"))
	}
	if s.TagKeys != nil && len(s.TagKeys) < 1 {
		invalidParams.Add(request.NewErrParamMinLen("TagKeys", 1))
	}

	if invalidParams.Len() > 0 {
		return invalidParams
	}
	return nil
}

// SetResourceArn sets the ResourceArn field's value.
func (s *UntagResourceInput) SetResourceArn(v string) *UntagResourceInput {
	s.ResourceArn = &v
	return s
}

// SetTagKeys sets the TagKeys field's value.
func (s *UntagResourceInput) SetTagKeys(v []*string) *UntagResourceInput {
	s.TagKeys = v
	return s
}

type UntagResourceOutput struct {
	_ struct{} `type:"structure"`
}

// String returns the string representation
func (s UntagResourceOutput) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s UntagResourceOutput) GoString() string {
	return s.String()
}
