// Code generated by smithy-go-codegen DO NOT EDIT.

package types

import (
	"fmt"
	smithy "github.com/aws/smithy-go"
)

// There is another ongoing conflicting backup control plane operation on the
// table. The backup is either being created, deleted or restored to a table.
type BackupInUseException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *BackupInUseException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *BackupInUseException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *BackupInUseException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "BackupInUseException"
	}
	return *e.ErrorCodeOverride
}
func (e *BackupInUseException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// Backup not found for the given BackupARN.
type BackupNotFoundException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *BackupNotFoundException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *BackupNotFoundException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *BackupNotFoundException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "BackupNotFoundException"
	}
	return *e.ErrorCodeOverride
}
func (e *BackupNotFoundException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// A condition specified in the operation failed to be evaluated.
type ConditionalCheckFailedException struct {
	Message *string

	ErrorCodeOverride *string

	Item map[string]AttributeValue

	noSmithyDocumentSerde
}

func (e *ConditionalCheckFailedException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *ConditionalCheckFailedException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *ConditionalCheckFailedException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "ConditionalCheckFailedException"
	}
	return *e.ErrorCodeOverride
}
func (e *ConditionalCheckFailedException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// Backups have not yet been enabled for this table.
type ContinuousBackupsUnavailableException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *ContinuousBackupsUnavailableException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *ContinuousBackupsUnavailableException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *ContinuousBackupsUnavailableException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "ContinuousBackupsUnavailableException"
	}
	return *e.ErrorCodeOverride
}
func (e *ContinuousBackupsUnavailableException) ErrorFault() smithy.ErrorFault {
	return smithy.FaultClient
}

//	There was an attempt to insert an item with the same primary key as an item
//
// that already exists in the DynamoDB table.
type DuplicateItemException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *DuplicateItemException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *DuplicateItemException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *DuplicateItemException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "DuplicateItemException"
	}
	return *e.ErrorCodeOverride
}
func (e *DuplicateItemException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// There was a conflict when writing to the specified S3 bucket.
type ExportConflictException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *ExportConflictException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *ExportConflictException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *ExportConflictException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "ExportConflictException"
	}
	return *e.ErrorCodeOverride
}
func (e *ExportConflictException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// The specified export was not found.
type ExportNotFoundException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *ExportNotFoundException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *ExportNotFoundException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *ExportNotFoundException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "ExportNotFoundException"
	}
	return *e.ErrorCodeOverride
}
func (e *ExportNotFoundException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// The specified global table already exists.
type GlobalTableAlreadyExistsException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *GlobalTableAlreadyExistsException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *GlobalTableAlreadyExistsException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *GlobalTableAlreadyExistsException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "GlobalTableAlreadyExistsException"
	}
	return *e.ErrorCodeOverride
}
func (e *GlobalTableAlreadyExistsException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// The specified global table does not exist.
type GlobalTableNotFoundException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *GlobalTableNotFoundException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *GlobalTableNotFoundException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *GlobalTableNotFoundException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "GlobalTableNotFoundException"
	}
	return *e.ErrorCodeOverride
}
func (e *GlobalTableNotFoundException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// DynamoDB rejected the request because you retried a request with a different
// payload but with an idempotent token that was already used.
type IdempotentParameterMismatchException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *IdempotentParameterMismatchException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *IdempotentParameterMismatchException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *IdempotentParameterMismatchException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "IdempotentParameterMismatchException"
	}
	return *e.ErrorCodeOverride
}
func (e *IdempotentParameterMismatchException) ErrorFault() smithy.ErrorFault {
	return smithy.FaultClient
}

//	There was a conflict when importing from the specified S3 source. This can
//
// occur when the current import conflicts with a previous import request that had
// the same client token.
type ImportConflictException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *ImportConflictException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *ImportConflictException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *ImportConflictException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "ImportConflictException"
	}
	return *e.ErrorCodeOverride
}
func (e *ImportConflictException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// The specified import was not found.
type ImportNotFoundException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *ImportNotFoundException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *ImportNotFoundException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *ImportNotFoundException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "ImportNotFoundException"
	}
	return *e.ErrorCodeOverride
}
func (e *ImportNotFoundException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// The operation tried to access a nonexistent index.
type IndexNotFoundException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *IndexNotFoundException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *IndexNotFoundException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *IndexNotFoundException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "IndexNotFoundException"
	}
	return *e.ErrorCodeOverride
}
func (e *IndexNotFoundException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// An error occurred on the server side.
type InternalServerError struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *InternalServerError) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *InternalServerError) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *InternalServerError) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "InternalServerError"
	}
	return *e.ErrorCodeOverride
}
func (e *InternalServerError) ErrorFault() smithy.ErrorFault { return smithy.FaultServer }

type InvalidEndpointException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *InvalidEndpointException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *InvalidEndpointException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *InvalidEndpointException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "InvalidEndpointException"
	}
	return *e.ErrorCodeOverride
}
func (e *InvalidEndpointException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// The specified ExportTime is outside of the point in time recovery window.
type InvalidExportTimeException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *InvalidExportTimeException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *InvalidExportTimeException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *InvalidExportTimeException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "InvalidExportTimeException"
	}
	return *e.ErrorCodeOverride
}
func (e *InvalidExportTimeException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// An invalid restore time was specified. RestoreDateTime must be between
// EarliestRestorableDateTime and LatestRestorableDateTime.
type InvalidRestoreTimeException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *InvalidRestoreTimeException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *InvalidRestoreTimeException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *InvalidRestoreTimeException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "InvalidRestoreTimeException"
	}
	return *e.ErrorCodeOverride
}
func (e *InvalidRestoreTimeException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// An item collection is too large. This exception is only returned for tables
// that have one or more local secondary indexes.
type ItemCollectionSizeLimitExceededException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *ItemCollectionSizeLimitExceededException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *ItemCollectionSizeLimitExceededException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *ItemCollectionSizeLimitExceededException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "ItemCollectionSizeLimitExceededException"
	}
	return *e.ErrorCodeOverride
}
func (e *ItemCollectionSizeLimitExceededException) ErrorFault() smithy.ErrorFault {
	return smithy.FaultClient
}

// There is no limit to the number of daily on-demand backups that can be taken.
//
// For most purposes, up to 500 simultaneous table operations are allowed per
// account. These operations include CreateTable , UpdateTable , DeleteTable ,
// UpdateTimeToLive , RestoreTableFromBackup , and RestoreTableToPointInTime .
//
// When you are creating a table with one or more secondary indexes, you can have
// up to 250 such requests running at a time. However, if the table or index
// specifications are complex, then DynamoDB might temporarily reduce the number of
// concurrent operations.
//
// When importing into DynamoDB, up to 50 simultaneous import table operations are
// allowed per account.
//
// There is a soft account quota of 2,500 tables.
//
// GetRecords was called with a value of more than 1000 for the limit request
// parameter.
//
// More than 2 processes are reading from the same streams shard at the same time.
// Exceeding this limit may result in request throttling.
type LimitExceededException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *LimitExceededException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *LimitExceededException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *LimitExceededException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "LimitExceededException"
	}
	return *e.ErrorCodeOverride
}
func (e *LimitExceededException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// Point in time recovery has not yet been enabled for this source table.
type PointInTimeRecoveryUnavailableException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *PointInTimeRecoveryUnavailableException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *PointInTimeRecoveryUnavailableException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *PointInTimeRecoveryUnavailableException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "PointInTimeRecoveryUnavailableException"
	}
	return *e.ErrorCodeOverride
}
func (e *PointInTimeRecoveryUnavailableException) ErrorFault() smithy.ErrorFault {
	return smithy.FaultClient
}

// The operation tried to access a nonexistent resource-based policy.
//
// If you specified an ExpectedRevisionId , it's possible that a policy is present
// for the resource but its revision ID didn't match the expected value.
type PolicyNotFoundException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *PolicyNotFoundException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *PolicyNotFoundException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *PolicyNotFoundException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "PolicyNotFoundException"
	}
	return *e.ErrorCodeOverride
}
func (e *PolicyNotFoundException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// Your request rate is too high. The Amazon Web Services SDKs for DynamoDB
// automatically retry requests that receive this exception. Your request is
// eventually successful, unless your retry queue is too large to finish. Reduce
// the frequency of requests and use exponential backoff. For more information, go
// to [Error Retries and Exponential Backoff]in the Amazon DynamoDB Developer Guide.
//
// [Error Retries and Exponential Backoff]: https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Programming.Errors.html#Programming.Errors.RetryAndBackoff
type ProvisionedThroughputExceededException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *ProvisionedThroughputExceededException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *ProvisionedThroughputExceededException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *ProvisionedThroughputExceededException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "ProvisionedThroughputExceededException"
	}
	return *e.ErrorCodeOverride
}
func (e *ProvisionedThroughputExceededException) ErrorFault() smithy.ErrorFault {
	return smithy.FaultClient
}

// The specified replica is already part of the global table.
type ReplicaAlreadyExistsException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *ReplicaAlreadyExistsException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *ReplicaAlreadyExistsException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *ReplicaAlreadyExistsException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "ReplicaAlreadyExistsException"
	}
	return *e.ErrorCodeOverride
}
func (e *ReplicaAlreadyExistsException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// The specified replica is no longer part of the global table.
type ReplicaNotFoundException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *ReplicaNotFoundException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *ReplicaNotFoundException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *ReplicaNotFoundException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "ReplicaNotFoundException"
	}
	return *e.ErrorCodeOverride
}
func (e *ReplicaNotFoundException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// The request was rejected because one or more items in the request are being
// modified by a request in another Region.
type ReplicatedWriteConflictException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *ReplicatedWriteConflictException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *ReplicatedWriteConflictException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *ReplicatedWriteConflictException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "ReplicatedWriteConflictException"
	}
	return *e.ErrorCodeOverride
}
func (e *ReplicatedWriteConflictException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// Throughput exceeds the current throughput quota for your account. Please
// contact [Amazon Web ServicesSupport]to request a quota increase.
//
// [Amazon Web ServicesSupport]: https://aws.amazon.com/support
type RequestLimitExceeded struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *RequestLimitExceeded) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *RequestLimitExceeded) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *RequestLimitExceeded) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "RequestLimitExceeded"
	}
	return *e.ErrorCodeOverride
}
func (e *RequestLimitExceeded) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// The operation conflicts with the resource's availability. For example:
//
//   - You attempted to recreate an existing table.
//
//   - You tried to delete a table currently in the CREATING state.
//
//   - You tried to update a resource that was already being updated.
//
// When appropriate, wait for the ongoing update to complete and attempt the
// request again.
type ResourceInUseException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *ResourceInUseException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *ResourceInUseException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *ResourceInUseException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "ResourceInUseException"
	}
	return *e.ErrorCodeOverride
}
func (e *ResourceInUseException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// The operation tried to access a nonexistent table or index. The resource might
// not be specified correctly, or its status might not be ACTIVE .
type ResourceNotFoundException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *ResourceNotFoundException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *ResourceNotFoundException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *ResourceNotFoundException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "ResourceNotFoundException"
	}
	return *e.ErrorCodeOverride
}
func (e *ResourceNotFoundException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// A target table with the specified name already exists.
type TableAlreadyExistsException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *TableAlreadyExistsException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *TableAlreadyExistsException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *TableAlreadyExistsException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "TableAlreadyExistsException"
	}
	return *e.ErrorCodeOverride
}
func (e *TableAlreadyExistsException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// A target table with the specified name is either being created or deleted.
type TableInUseException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *TableInUseException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *TableInUseException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *TableInUseException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "TableInUseException"
	}
	return *e.ErrorCodeOverride
}
func (e *TableInUseException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// A source table with the name TableName does not currently exist within the
// subscriber's account or the subscriber is operating in the wrong Amazon Web
// Services Region.
type TableNotFoundException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *TableNotFoundException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *TableNotFoundException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *TableNotFoundException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "TableNotFoundException"
	}
	return *e.ErrorCodeOverride
}
func (e *TableNotFoundException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// The entire transaction request was canceled.
//
// DynamoDB cancels a TransactWriteItems request under the following circumstances:
//
//   - A condition in one of the condition expressions is not met.
//
//   - A table in the TransactWriteItems request is in a different account or
//     region.
//
//   - More than one action in the TransactWriteItems operation targets the same
//     item.
//
//   - There is insufficient provisioned capacity for the transaction to be
//     completed.
//
//   - An item size becomes too large (larger than 400 KB), or a local secondary
//     index (LSI) becomes too large, or a similar validation error occurs because of
//     changes made by the transaction.
//
//   - There is a user error, such as an invalid data format.
//
//   - There is an ongoing TransactWriteItems operation that conflicts with a
//     concurrent TransactWriteItems request. In this case the TransactWriteItems
//     operation fails with a TransactionCanceledException .
//
// DynamoDB cancels a TransactGetItems request under the following circumstances:
//
//   - There is an ongoing TransactGetItems operation that conflicts with a
//     concurrent PutItem , UpdateItem , DeleteItem or TransactWriteItems request. In
//     this case the TransactGetItems operation fails with a
//     TransactionCanceledException .
//
//   - A table in the TransactGetItems request is in a different account or region.
//
//   - There is insufficient provisioned capacity for the transaction to be
//     completed.
//
//   - There is a user error, such as an invalid data format.
//
// If using Java, DynamoDB lists the cancellation reasons on the
// CancellationReasons property. This property is not set for other languages.
// Transaction cancellation reasons are ordered in the order of requested items, if
// an item has no error it will have None code and Null message.
//
// Cancellation reason codes and possible error messages:
//
//   - No Errors:
//
//   - Code: None
//
//   - Message: null
//
//   - Conditional Check Failed:
//
//   - Code: ConditionalCheckFailed
//
//   - Message: The conditional request failed.
//
//   - Item Collection Size Limit Exceeded:
//
//   - Code: ItemCollectionSizeLimitExceeded
//
//   - Message: Collection size exceeded.
//
//   - Transaction Conflict:
//
//   - Code: TransactionConflict
//
//   - Message: Transaction is ongoing for the item.
//
//   - Provisioned Throughput Exceeded:
//
//   - Code: ProvisionedThroughputExceeded
//
//   - Messages:
//
//   - The level of configured provisioned throughput for the table was exceeded.
//     Consider increasing your provisioning level with the UpdateTable API.
//
// This Message is received when provisioned throughput is exceeded is on a
//
//	provisioned DynamoDB table.
//
//	- The level of configured provisioned throughput for one or more global
//	secondary indexes of the table was exceeded. Consider increasing your
//	provisioning level for the under-provisioned global secondary indexes with the
//	UpdateTable API.
//
// This message is returned when provisioned throughput is exceeded is on a
//
//	provisioned GSI.
//
//	- Throttling Error:
//
//	- Code: ThrottlingError
//
//	- Messages:
//
//	- Throughput exceeds the current capacity of your table or index. DynamoDB is
//	automatically scaling your table or index so please try again shortly. If
//	exceptions persist, check if you have a hot key:
//	https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/bp-partition-key-design.html.
//
// This message is returned when writes get throttled on an On-Demand table as
//
//	DynamoDB is automatically scaling the table.
//
//	- Throughput exceeds the current capacity for one or more global secondary
//	indexes. DynamoDB is automatically scaling your index so please try again
//	shortly.
//
// This message is returned when writes get throttled on an On-Demand GSI as
//
//	DynamoDB is automatically scaling the GSI.
//
//	- Validation Error:
//
//	- Code: ValidationError
//
//	- Messages:
//
//	- One or more parameter values were invalid.
//
//	- The update expression attempted to update the secondary index key beyond
//	allowed size limits.
//
//	- The update expression attempted to update the secondary index key to
//	unsupported type.
//
//	- An operand in the update expression has an incorrect data type.
//
//	- Item size to update has exceeded the maximum allowed size.
//
//	- Number overflow. Attempting to store a number with magnitude larger than
//	supported range.
//
//	- Type mismatch for attribute to update.
//
//	- Nesting Levels have exceeded supported limits.
//
//	- The document path provided in the update expression is invalid for update.
//
//	- The provided expression refers to an attribute that does not exist in the
//	item.
type TransactionCanceledException struct {
	Message *string

	ErrorCodeOverride *string

	CancellationReasons []CancellationReason

	noSmithyDocumentSerde
}

func (e *TransactionCanceledException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *TransactionCanceledException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *TransactionCanceledException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "TransactionCanceledException"
	}
	return *e.ErrorCodeOverride
}
func (e *TransactionCanceledException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// Operation was rejected because there is an ongoing transaction for the item.
type TransactionConflictException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *TransactionConflictException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *TransactionConflictException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *TransactionConflictException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "TransactionConflictException"
	}
	return *e.ErrorCodeOverride
}
func (e *TransactionConflictException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// The transaction with the given request token is already in progress.
//
// # Recommended Settings
//
// This is a general recommendation for handling the TransactionInProgressException
// . These settings help ensure that the client retries will trigger completion of
// the ongoing TransactWriteItems request.
//
//   - Set clientExecutionTimeout to a value that allows at least one retry to be
//     processed after 5 seconds have elapsed since the first attempt for the
//     TransactWriteItems operation.
//
//   - Set socketTimeout to a value a little lower than the requestTimeout setting.
//
//   - requestTimeout should be set based on the time taken for the individual
//     retries of a single HTTP request for your use case, but setting it to 1 second
//     or higher should work well to reduce chances of retries and
//     TransactionInProgressException errors.
//
//   - Use exponential backoff when retrying and tune backoff if needed.
//
// Assuming [default retry policy], example timeout settings based on the guidelines above are as
// follows:
//
// Example timeline:
//
//   - 0-1000 first attempt
//
//   - 1000-1500 first sleep/delay (default retry policy uses 500 ms as base delay
//     for 4xx errors)
//
//   - 1500-2500 second attempt
//
//   - 2500-3500 second sleep/delay (500 * 2, exponential backoff)
//
//   - 3500-4500 third attempt
//
//   - 4500-6500 third sleep/delay (500 * 2^2)
//
//   - 6500-7500 fourth attempt (this can trigger inline recovery since 5 seconds
//     have elapsed since the first attempt reached TC)
//
// [default retry policy]: https://github.com/aws/aws-sdk-java/blob/fd409dee8ae23fb8953e0bb4dbde65536a7e0514/aws-java-sdk-core/src/main/java/com/amazonaws/retry/PredefinedRetryPolicies.java#L97
type TransactionInProgressException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *TransactionInProgressException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *TransactionInProgressException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *TransactionInProgressException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "TransactionInProgressException"
	}
	return *e.ErrorCodeOverride
}
func (e *TransactionInProgressException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }
