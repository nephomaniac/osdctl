// Code generated by MockGen. DO NOT EDIT.
// Source: client.go

// Package mock is a generated GoMock package.
package mock

import (
	reflect "reflect"

	cloudtrail "github.com/aws/aws-sdk-go/service/cloudtrail"
	costexplorer "github.com/aws/aws-sdk-go/service/costexplorer"
	ec2 "github.com/aws/aws-sdk-go/service/ec2"
	iam "github.com/aws/aws-sdk-go/service/iam"
	organizations "github.com/aws/aws-sdk-go/service/organizations"
	resourcegroupstaggingapi "github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi"
	s3 "github.com/aws/aws-sdk-go/service/s3"
	servicequotas "github.com/aws/aws-sdk-go/service/servicequotas"
	sts "github.com/aws/aws-sdk-go/service/sts"
	gomock "github.com/golang/mock/gomock"
)

// MockClient is a mock of Client interface.
type MockClient struct {
	ctrl     *gomock.Controller
	recorder *MockClientMockRecorder
}

// MockClientMockRecorder is the mock recorder for MockClient.
type MockClientMockRecorder struct {
	mock *MockClient
}

// NewMockClient creates a new mock instance.
func NewMockClient(ctrl *gomock.Controller) *MockClient {
	mock := &MockClient{ctrl: ctrl}
	mock.recorder = &MockClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockClient) EXPECT() *MockClientMockRecorder {
	return m.recorder
}

// AssumeRole mocks base method.
func (m *MockClient) AssumeRole(arg0 *sts.AssumeRoleInput) (*sts.AssumeRoleOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AssumeRole", arg0)
	ret0, _ := ret[0].(*sts.AssumeRoleOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AssumeRole indicates an expected call of AssumeRole.
func (mr *MockClientMockRecorder) AssumeRole(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AssumeRole", reflect.TypeOf((*MockClient)(nil).AssumeRole), arg0)
}

// AttachRolePolicy mocks base method.
func (m *MockClient) AttachRolePolicy(arg0 *iam.AttachRolePolicyInput) (*iam.AttachRolePolicyOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AttachRolePolicy", arg0)
	ret0, _ := ret[0].(*iam.AttachRolePolicyOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AttachRolePolicy indicates an expected call of AttachRolePolicy.
func (mr *MockClientMockRecorder) AttachRolePolicy(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AttachRolePolicy", reflect.TypeOf((*MockClient)(nil).AttachRolePolicy), arg0)
}

// AttachUserPolicy mocks base method.
func (m *MockClient) AttachUserPolicy(arg0 *iam.AttachUserPolicyInput) (*iam.AttachUserPolicyOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AttachUserPolicy", arg0)
	ret0, _ := ret[0].(*iam.AttachUserPolicyOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AttachUserPolicy indicates an expected call of AttachUserPolicy.
func (mr *MockClientMockRecorder) AttachUserPolicy(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AttachUserPolicy", reflect.TypeOf((*MockClient)(nil).AttachUserPolicy), arg0)
}

// CreateAccessKey mocks base method.
func (m *MockClient) CreateAccessKey(arg0 *iam.CreateAccessKeyInput) (*iam.CreateAccessKeyOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateAccessKey", arg0)
	ret0, _ := ret[0].(*iam.CreateAccessKeyOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateAccessKey indicates an expected call of CreateAccessKey.
func (mr *MockClientMockRecorder) CreateAccessKey(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateAccessKey", reflect.TypeOf((*MockClient)(nil).CreateAccessKey), arg0)
}

// CreateAccount mocks base method.
func (m *MockClient) CreateAccount(input *organizations.CreateAccountInput) (*organizations.CreateAccountOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateAccount", input)
	ret0, _ := ret[0].(*organizations.CreateAccountOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateAccount indicates an expected call of CreateAccount.
func (mr *MockClientMockRecorder) CreateAccount(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateAccount", reflect.TypeOf((*MockClient)(nil).CreateAccount), input)
}

// CreateCostCategoryDefinition mocks base method.
func (m *MockClient) CreateCostCategoryDefinition(input *costexplorer.CreateCostCategoryDefinitionInput) (*costexplorer.CreateCostCategoryDefinitionOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateCostCategoryDefinition", input)
	ret0, _ := ret[0].(*costexplorer.CreateCostCategoryDefinitionOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateCostCategoryDefinition indicates an expected call of CreateCostCategoryDefinition.
func (mr *MockClientMockRecorder) CreateCostCategoryDefinition(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateCostCategoryDefinition", reflect.TypeOf((*MockClient)(nil).CreateCostCategoryDefinition), input)
}

// CreatePolicy mocks base method.
func (m *MockClient) CreatePolicy(arg0 *iam.CreatePolicyInput) (*iam.CreatePolicyOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreatePolicy", arg0)
	ret0, _ := ret[0].(*iam.CreatePolicyOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreatePolicy indicates an expected call of CreatePolicy.
func (mr *MockClientMockRecorder) CreatePolicy(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreatePolicy", reflect.TypeOf((*MockClient)(nil).CreatePolicy), arg0)
}

// CreateUser mocks base method.
func (m *MockClient) CreateUser(arg0 *iam.CreateUserInput) (*iam.CreateUserOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateUser", arg0)
	ret0, _ := ret[0].(*iam.CreateUserOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateUser indicates an expected call of CreateUser.
func (mr *MockClientMockRecorder) CreateUser(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateUser", reflect.TypeOf((*MockClient)(nil).CreateUser), arg0)
}

// DeleteAccessKey mocks base method.
func (m *MockClient) DeleteAccessKey(arg0 *iam.DeleteAccessKeyInput) (*iam.DeleteAccessKeyOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteAccessKey", arg0)
	ret0, _ := ret[0].(*iam.DeleteAccessKeyOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteAccessKey indicates an expected call of DeleteAccessKey.
func (mr *MockClientMockRecorder) DeleteAccessKey(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteAccessKey", reflect.TypeOf((*MockClient)(nil).DeleteAccessKey), arg0)
}

// DeleteBucket mocks base method.
func (m *MockClient) DeleteBucket(arg0 *s3.DeleteBucketInput) (*s3.DeleteBucketOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteBucket", arg0)
	ret0, _ := ret[0].(*s3.DeleteBucketOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteBucket indicates an expected call of DeleteBucket.
func (mr *MockClientMockRecorder) DeleteBucket(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteBucket", reflect.TypeOf((*MockClient)(nil).DeleteBucket), arg0)
}

// DeleteLoginProfile mocks base method.
func (m *MockClient) DeleteLoginProfile(arg0 *iam.DeleteLoginProfileInput) (*iam.DeleteLoginProfileOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteLoginProfile", arg0)
	ret0, _ := ret[0].(*iam.DeleteLoginProfileOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteLoginProfile indicates an expected call of DeleteLoginProfile.
func (mr *MockClientMockRecorder) DeleteLoginProfile(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteLoginProfile", reflect.TypeOf((*MockClient)(nil).DeleteLoginProfile), arg0)
}

// DeleteObjects mocks base method.
func (m *MockClient) DeleteObjects(arg0 *s3.DeleteObjectsInput) (*s3.DeleteObjectsOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteObjects", arg0)
	ret0, _ := ret[0].(*s3.DeleteObjectsOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteObjects indicates an expected call of DeleteObjects.
func (mr *MockClientMockRecorder) DeleteObjects(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteObjects", reflect.TypeOf((*MockClient)(nil).DeleteObjects), arg0)
}

// DeletePolicy mocks base method.
func (m *MockClient) DeletePolicy(arg0 *iam.DeletePolicyInput) (*iam.DeletePolicyOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeletePolicy", arg0)
	ret0, _ := ret[0].(*iam.DeletePolicyOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeletePolicy indicates an expected call of DeletePolicy.
func (mr *MockClientMockRecorder) DeletePolicy(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeletePolicy", reflect.TypeOf((*MockClient)(nil).DeletePolicy), arg0)
}

// DeleteRole mocks base method.
func (m *MockClient) DeleteRole(arg0 *iam.DeleteRoleInput) (*iam.DeleteRoleOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteRole", arg0)
	ret0, _ := ret[0].(*iam.DeleteRoleOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteRole indicates an expected call of DeleteRole.
func (mr *MockClientMockRecorder) DeleteRole(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteRole", reflect.TypeOf((*MockClient)(nil).DeleteRole), arg0)
}

// DeleteSigningCertificate mocks base method.
func (m *MockClient) DeleteSigningCertificate(arg0 *iam.DeleteSigningCertificateInput) (*iam.DeleteSigningCertificateOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteSigningCertificate", arg0)
	ret0, _ := ret[0].(*iam.DeleteSigningCertificateOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteSigningCertificate indicates an expected call of DeleteSigningCertificate.
func (mr *MockClientMockRecorder) DeleteSigningCertificate(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteSigningCertificate", reflect.TypeOf((*MockClient)(nil).DeleteSigningCertificate), arg0)
}

// DeleteUser mocks base method.
func (m *MockClient) DeleteUser(arg0 *iam.DeleteUserInput) (*iam.DeleteUserOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteUser", arg0)
	ret0, _ := ret[0].(*iam.DeleteUserOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteUser indicates an expected call of DeleteUser.
func (mr *MockClientMockRecorder) DeleteUser(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteUser", reflect.TypeOf((*MockClient)(nil).DeleteUser), arg0)
}

// DeleteUserPolicy mocks base method.
func (m *MockClient) DeleteUserPolicy(arg0 *iam.DeleteUserPolicyInput) (*iam.DeleteUserPolicyOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteUserPolicy", arg0)
	ret0, _ := ret[0].(*iam.DeleteUserPolicyOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteUserPolicy indicates an expected call of DeleteUserPolicy.
func (mr *MockClientMockRecorder) DeleteUserPolicy(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteUserPolicy", reflect.TypeOf((*MockClient)(nil).DeleteUserPolicy), arg0)
}

// DescribeAccount mocks base method.
func (m *MockClient) DescribeAccount(input *organizations.DescribeAccountInput) (*organizations.DescribeAccountOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DescribeAccount", input)
	ret0, _ := ret[0].(*organizations.DescribeAccountOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DescribeAccount indicates an expected call of DescribeAccount.
func (mr *MockClientMockRecorder) DescribeAccount(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DescribeAccount", reflect.TypeOf((*MockClient)(nil).DescribeAccount), input)
}

// DescribeCreateAccountStatus mocks base method.
func (m *MockClient) DescribeCreateAccountStatus(input *organizations.DescribeCreateAccountStatusInput) (*organizations.DescribeCreateAccountStatusOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DescribeCreateAccountStatus", input)
	ret0, _ := ret[0].(*organizations.DescribeCreateAccountStatusOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DescribeCreateAccountStatus indicates an expected call of DescribeCreateAccountStatus.
func (mr *MockClientMockRecorder) DescribeCreateAccountStatus(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DescribeCreateAccountStatus", reflect.TypeOf((*MockClient)(nil).DescribeCreateAccountStatus), input)
}

// DescribeInstances mocks base method.
func (m *MockClient) DescribeInstances(arg0 *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DescribeInstances", arg0)
	ret0, _ := ret[0].(*ec2.DescribeInstancesOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DescribeInstances indicates an expected call of DescribeInstances.
func (mr *MockClientMockRecorder) DescribeInstances(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DescribeInstances", reflect.TypeOf((*MockClient)(nil).DescribeInstances), arg0)
}

// DescribeOrganizationalUnit mocks base method.
func (m *MockClient) DescribeOrganizationalUnit(input *organizations.DescribeOrganizationalUnitInput) (*organizations.DescribeOrganizationalUnitOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DescribeOrganizationalUnit", input)
	ret0, _ := ret[0].(*organizations.DescribeOrganizationalUnitOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DescribeOrganizationalUnit indicates an expected call of DescribeOrganizationalUnit.
func (mr *MockClientMockRecorder) DescribeOrganizationalUnit(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DescribeOrganizationalUnit", reflect.TypeOf((*MockClient)(nil).DescribeOrganizationalUnit), input)
}

// DescribeRouteTables mocks base method.
func (m *MockClient) DescribeRouteTables(arg0 *ec2.DescribeRouteTablesInput) (*ec2.DescribeRouteTablesOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DescribeRouteTables", arg0)
	ret0, _ := ret[0].(*ec2.DescribeRouteTablesOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DescribeRouteTables indicates an expected call of DescribeRouteTables.
func (mr *MockClientMockRecorder) DescribeRouteTables(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DescribeRouteTables", reflect.TypeOf((*MockClient)(nil).DescribeRouteTables), arg0)
}

// DescribeSubnets mocks base method.
func (m *MockClient) DescribeSubnets(arg0 *ec2.DescribeSubnetsInput) (*ec2.DescribeSubnetsOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DescribeSubnets", arg0)
	ret0, _ := ret[0].(*ec2.DescribeSubnetsOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DescribeSubnets indicates an expected call of DescribeSubnets.
func (mr *MockClientMockRecorder) DescribeSubnets(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DescribeSubnets", reflect.TypeOf((*MockClient)(nil).DescribeSubnets), arg0)
}

// DescribeVpcs mocks base method.
func (m *MockClient) DescribeVpcs(arg0 *ec2.DescribeVpcsInput) (*ec2.DescribeVpcsOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DescribeVpcs", arg0)
	ret0, _ := ret[0].(*ec2.DescribeVpcsOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DescribeVpcs indicates an expected call of DescribeVpcs.
func (mr *MockClientMockRecorder) DescribeVpcs(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DescribeVpcs", reflect.TypeOf((*MockClient)(nil).DescribeVpcs), arg0)
}

// DetachRolePolicy mocks base method.
func (m *MockClient) DetachRolePolicy(arg0 *iam.DetachRolePolicyInput) (*iam.DetachRolePolicyOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DetachRolePolicy", arg0)
	ret0, _ := ret[0].(*iam.DetachRolePolicyOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DetachRolePolicy indicates an expected call of DetachRolePolicy.
func (mr *MockClientMockRecorder) DetachRolePolicy(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DetachRolePolicy", reflect.TypeOf((*MockClient)(nil).DetachRolePolicy), arg0)
}

// DetachUserPolicy mocks base method.
func (m *MockClient) DetachUserPolicy(arg0 *iam.DetachUserPolicyInput) (*iam.DetachUserPolicyOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DetachUserPolicy", arg0)
	ret0, _ := ret[0].(*iam.DetachUserPolicyOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DetachUserPolicy indicates an expected call of DetachUserPolicy.
func (mr *MockClientMockRecorder) DetachUserPolicy(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DetachUserPolicy", reflect.TypeOf((*MockClient)(nil).DetachUserPolicy), arg0)
}

// GetCallerIdentity mocks base method.
func (m *MockClient) GetCallerIdentity(arg0 *sts.GetCallerIdentityInput) (*sts.GetCallerIdentityOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCallerIdentity", arg0)
	ret0, _ := ret[0].(*sts.GetCallerIdentityOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCallerIdentity indicates an expected call of GetCallerIdentity.
func (mr *MockClientMockRecorder) GetCallerIdentity(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCallerIdentity", reflect.TypeOf((*MockClient)(nil).GetCallerIdentity), arg0)
}

// GetCostAndUsage mocks base method.
func (m *MockClient) GetCostAndUsage(input *costexplorer.GetCostAndUsageInput) (*costexplorer.GetCostAndUsageOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCostAndUsage", input)
	ret0, _ := ret[0].(*costexplorer.GetCostAndUsageOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCostAndUsage indicates an expected call of GetCostAndUsage.
func (mr *MockClientMockRecorder) GetCostAndUsage(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCostAndUsage", reflect.TypeOf((*MockClient)(nil).GetCostAndUsage), input)
}

// GetFederationToken mocks base method.
func (m *MockClient) GetFederationToken(arg0 *sts.GetFederationTokenInput) (*sts.GetFederationTokenOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetFederationToken", arg0)
	ret0, _ := ret[0].(*sts.GetFederationTokenOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetFederationToken indicates an expected call of GetFederationToken.
func (mr *MockClientMockRecorder) GetFederationToken(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetFederationToken", reflect.TypeOf((*MockClient)(nil).GetFederationToken), arg0)
}

// GetResources mocks base method.
func (m *MockClient) GetResources(input *resourcegroupstaggingapi.GetResourcesInput) (*resourcegroupstaggingapi.GetResourcesOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetResources", input)
	ret0, _ := ret[0].(*resourcegroupstaggingapi.GetResourcesOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetResources indicates an expected call of GetResources.
func (mr *MockClientMockRecorder) GetResources(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetResources", reflect.TypeOf((*MockClient)(nil).GetResources), input)
}

// GetUser mocks base method.
func (m *MockClient) GetUser(arg0 *iam.GetUserInput) (*iam.GetUserOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUser", arg0)
	ret0, _ := ret[0].(*iam.GetUserOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUser indicates an expected call of GetUser.
func (mr *MockClientMockRecorder) GetUser(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUser", reflect.TypeOf((*MockClient)(nil).GetUser), arg0)
}

// ListAccessKeys mocks base method.
func (m *MockClient) ListAccessKeys(arg0 *iam.ListAccessKeysInput) (*iam.ListAccessKeysOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListAccessKeys", arg0)
	ret0, _ := ret[0].(*iam.ListAccessKeysOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListAccessKeys indicates an expected call of ListAccessKeys.
func (mr *MockClientMockRecorder) ListAccessKeys(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListAccessKeys", reflect.TypeOf((*MockClient)(nil).ListAccessKeys), arg0)
}

// ListAccounts mocks base method.
func (m *MockClient) ListAccounts(input *organizations.ListAccountsInput) (*organizations.ListAccountsOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListAccounts", input)
	ret0, _ := ret[0].(*organizations.ListAccountsOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListAccounts indicates an expected call of ListAccounts.
func (mr *MockClientMockRecorder) ListAccounts(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListAccounts", reflect.TypeOf((*MockClient)(nil).ListAccounts), input)
}

// ListAccountsForParent mocks base method.
func (m *MockClient) ListAccountsForParent(input *organizations.ListAccountsForParentInput) (*organizations.ListAccountsForParentOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListAccountsForParent", input)
	ret0, _ := ret[0].(*organizations.ListAccountsForParentOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListAccountsForParent indicates an expected call of ListAccountsForParent.
func (mr *MockClientMockRecorder) ListAccountsForParent(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListAccountsForParent", reflect.TypeOf((*MockClient)(nil).ListAccountsForParent), input)
}

// ListAttachedRolePolicies mocks base method.
func (m *MockClient) ListAttachedRolePolicies(arg0 *iam.ListAttachedRolePoliciesInput) (*iam.ListAttachedRolePoliciesOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListAttachedRolePolicies", arg0)
	ret0, _ := ret[0].(*iam.ListAttachedRolePoliciesOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListAttachedRolePolicies indicates an expected call of ListAttachedRolePolicies.
func (mr *MockClientMockRecorder) ListAttachedRolePolicies(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListAttachedRolePolicies", reflect.TypeOf((*MockClient)(nil).ListAttachedRolePolicies), arg0)
}

// ListAttachedUserPolicies mocks base method.
func (m *MockClient) ListAttachedUserPolicies(arg0 *iam.ListAttachedUserPoliciesInput) (*iam.ListAttachedUserPoliciesOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListAttachedUserPolicies", arg0)
	ret0, _ := ret[0].(*iam.ListAttachedUserPoliciesOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListAttachedUserPolicies indicates an expected call of ListAttachedUserPolicies.
func (mr *MockClientMockRecorder) ListAttachedUserPolicies(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListAttachedUserPolicies", reflect.TypeOf((*MockClient)(nil).ListAttachedUserPolicies), arg0)
}

// ListBuckets mocks base method.
func (m *MockClient) ListBuckets(arg0 *s3.ListBucketsInput) (*s3.ListBucketsOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListBuckets", arg0)
	ret0, _ := ret[0].(*s3.ListBucketsOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListBuckets indicates an expected call of ListBuckets.
func (mr *MockClientMockRecorder) ListBuckets(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListBuckets", reflect.TypeOf((*MockClient)(nil).ListBuckets), arg0)
}

// ListChildren mocks base method.
func (m *MockClient) ListChildren(input *organizations.ListChildrenInput) (*organizations.ListChildrenOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListChildren", input)
	ret0, _ := ret[0].(*organizations.ListChildrenOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListChildren indicates an expected call of ListChildren.
func (mr *MockClientMockRecorder) ListChildren(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListChildren", reflect.TypeOf((*MockClient)(nil).ListChildren), input)
}

// ListCostCategoryDefinitions mocks base method.
func (m *MockClient) ListCostCategoryDefinitions(input *costexplorer.ListCostCategoryDefinitionsInput) (*costexplorer.ListCostCategoryDefinitionsOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListCostCategoryDefinitions", input)
	ret0, _ := ret[0].(*costexplorer.ListCostCategoryDefinitionsOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListCostCategoryDefinitions indicates an expected call of ListCostCategoryDefinitions.
func (mr *MockClientMockRecorder) ListCostCategoryDefinitions(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListCostCategoryDefinitions", reflect.TypeOf((*MockClient)(nil).ListCostCategoryDefinitions), input)
}

// ListGroupsForUser mocks base method.
func (m *MockClient) ListGroupsForUser(arg0 *iam.ListGroupsForUserInput) (*iam.ListGroupsForUserOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListGroupsForUser", arg0)
	ret0, _ := ret[0].(*iam.ListGroupsForUserOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListGroupsForUser indicates an expected call of ListGroupsForUser.
func (mr *MockClientMockRecorder) ListGroupsForUser(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListGroupsForUser", reflect.TypeOf((*MockClient)(nil).ListGroupsForUser), arg0)
}

// ListObjects mocks base method.
func (m *MockClient) ListObjects(arg0 *s3.ListObjectsInput) (*s3.ListObjectsOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListObjects", arg0)
	ret0, _ := ret[0].(*s3.ListObjectsOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListObjects indicates an expected call of ListObjects.
func (mr *MockClientMockRecorder) ListObjects(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListObjects", reflect.TypeOf((*MockClient)(nil).ListObjects), arg0)
}

// ListOrganizationalUnitsForParent mocks base method.
func (m *MockClient) ListOrganizationalUnitsForParent(input *organizations.ListOrganizationalUnitsForParentInput) (*organizations.ListOrganizationalUnitsForParentOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListOrganizationalUnitsForParent", input)
	ret0, _ := ret[0].(*organizations.ListOrganizationalUnitsForParentOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListOrganizationalUnitsForParent indicates an expected call of ListOrganizationalUnitsForParent.
func (mr *MockClientMockRecorder) ListOrganizationalUnitsForParent(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListOrganizationalUnitsForParent", reflect.TypeOf((*MockClient)(nil).ListOrganizationalUnitsForParent), input)
}

// ListParents mocks base method.
func (m *MockClient) ListParents(input *organizations.ListParentsInput) (*organizations.ListParentsOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListParents", input)
	ret0, _ := ret[0].(*organizations.ListParentsOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListParents indicates an expected call of ListParents.
func (mr *MockClientMockRecorder) ListParents(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListParents", reflect.TypeOf((*MockClient)(nil).ListParents), input)
}

// ListPolicies mocks base method.
func (m *MockClient) ListPolicies(arg0 *iam.ListPoliciesInput) (*iam.ListPoliciesOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListPolicies", arg0)
	ret0, _ := ret[0].(*iam.ListPoliciesOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListPolicies indicates an expected call of ListPolicies.
func (mr *MockClientMockRecorder) ListPolicies(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListPolicies", reflect.TypeOf((*MockClient)(nil).ListPolicies), arg0)
}

// ListRoles mocks base method.
func (m *MockClient) ListRoles(arg0 *iam.ListRolesInput) (*iam.ListRolesOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListRoles", arg0)
	ret0, _ := ret[0].(*iam.ListRolesOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListRoles indicates an expected call of ListRoles.
func (mr *MockClientMockRecorder) ListRoles(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListRoles", reflect.TypeOf((*MockClient)(nil).ListRoles), arg0)
}

// ListRoots mocks base method.
func (m *MockClient) ListRoots(input *organizations.ListRootsInput) (*organizations.ListRootsOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListRoots", input)
	ret0, _ := ret[0].(*organizations.ListRootsOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListRoots indicates an expected call of ListRoots.
func (mr *MockClientMockRecorder) ListRoots(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListRoots", reflect.TypeOf((*MockClient)(nil).ListRoots), input)
}

// ListServiceQuotas mocks base method.
func (m *MockClient) ListServiceQuotas(arg0 *servicequotas.ListServiceQuotasInput) (*servicequotas.ListServiceQuotasOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListServiceQuotas", arg0)
	ret0, _ := ret[0].(*servicequotas.ListServiceQuotasOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListServiceQuotas indicates an expected call of ListServiceQuotas.
func (mr *MockClientMockRecorder) ListServiceQuotas(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListServiceQuotas", reflect.TypeOf((*MockClient)(nil).ListServiceQuotas), arg0)
}

// ListSigningCertificates mocks base method.
func (m *MockClient) ListSigningCertificates(arg0 *iam.ListSigningCertificatesInput) (*iam.ListSigningCertificatesOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListSigningCertificates", arg0)
	ret0, _ := ret[0].(*iam.ListSigningCertificatesOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListSigningCertificates indicates an expected call of ListSigningCertificates.
func (mr *MockClientMockRecorder) ListSigningCertificates(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListSigningCertificates", reflect.TypeOf((*MockClient)(nil).ListSigningCertificates), arg0)
}

// ListTagsForResource mocks base method.
func (m *MockClient) ListTagsForResource(input *organizations.ListTagsForResourceInput) (*organizations.ListTagsForResourceOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListTagsForResource", input)
	ret0, _ := ret[0].(*organizations.ListTagsForResourceOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListTagsForResource indicates an expected call of ListTagsForResource.
func (mr *MockClientMockRecorder) ListTagsForResource(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListTagsForResource", reflect.TypeOf((*MockClient)(nil).ListTagsForResource), input)
}

// ListUserPolicies mocks base method.
func (m *MockClient) ListUserPolicies(arg0 *iam.ListUserPoliciesInput) (*iam.ListUserPoliciesOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListUserPolicies", arg0)
	ret0, _ := ret[0].(*iam.ListUserPoliciesOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListUserPolicies indicates an expected call of ListUserPolicies.
func (mr *MockClientMockRecorder) ListUserPolicies(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListUserPolicies", reflect.TypeOf((*MockClient)(nil).ListUserPolicies), arg0)
}

// ListUsers mocks base method.
func (m *MockClient) ListUsers(arg0 *iam.ListUsersInput) (*iam.ListUsersOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListUsers", arg0)
	ret0, _ := ret[0].(*iam.ListUsersOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListUsers indicates an expected call of ListUsers.
func (mr *MockClientMockRecorder) ListUsers(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListUsers", reflect.TypeOf((*MockClient)(nil).ListUsers), arg0)
}

// LookupEvents mocks base method.
func (m *MockClient) LookupEvents(input *cloudtrail.LookupEventsInput) (*cloudtrail.LookupEventsOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LookupEvents", input)
	ret0, _ := ret[0].(*cloudtrail.LookupEventsOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// LookupEvents indicates an expected call of LookupEvents.
func (mr *MockClientMockRecorder) LookupEvents(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LookupEvents", reflect.TypeOf((*MockClient)(nil).LookupEvents), input)
}

// MoveAccount mocks base method.
func (m *MockClient) MoveAccount(input *organizations.MoveAccountInput) (*organizations.MoveAccountOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MoveAccount", input)
	ret0, _ := ret[0].(*organizations.MoveAccountOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MoveAccount indicates an expected call of MoveAccount.
func (mr *MockClientMockRecorder) MoveAccount(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MoveAccount", reflect.TypeOf((*MockClient)(nil).MoveAccount), input)
}

// RemoveUserFromGroup mocks base method.
func (m *MockClient) RemoveUserFromGroup(arg0 *iam.RemoveUserFromGroupInput) (*iam.RemoveUserFromGroupOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoveUserFromGroup", arg0)
	ret0, _ := ret[0].(*iam.RemoveUserFromGroupOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RemoveUserFromGroup indicates an expected call of RemoveUserFromGroup.
func (mr *MockClientMockRecorder) RemoveUserFromGroup(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveUserFromGroup", reflect.TypeOf((*MockClient)(nil).RemoveUserFromGroup), arg0)
}

// RequestServiceQuotaIncrease mocks base method.
func (m *MockClient) RequestServiceQuotaIncrease(arg0 *servicequotas.RequestServiceQuotaIncreaseInput) (*servicequotas.RequestServiceQuotaIncreaseOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RequestServiceQuotaIncrease", arg0)
	ret0, _ := ret[0].(*servicequotas.RequestServiceQuotaIncreaseOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RequestServiceQuotaIncrease indicates an expected call of RequestServiceQuotaIncrease.
func (mr *MockClientMockRecorder) RequestServiceQuotaIncrease(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RequestServiceQuotaIncrease", reflect.TypeOf((*MockClient)(nil).RequestServiceQuotaIncrease), arg0)
}

// TagResource mocks base method.
func (m *MockClient) TagResource(input *organizations.TagResourceInput) (*organizations.TagResourceOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TagResource", input)
	ret0, _ := ret[0].(*organizations.TagResourceOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// TagResource indicates an expected call of TagResource.
func (mr *MockClientMockRecorder) TagResource(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TagResource", reflect.TypeOf((*MockClient)(nil).TagResource), input)
}

// UntagResource mocks base method.
func (m *MockClient) UntagResource(input *organizations.UntagResourceInput) (*organizations.UntagResourceOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UntagResource", input)
	ret0, _ := ret[0].(*organizations.UntagResourceOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UntagResource indicates an expected call of UntagResource.
func (mr *MockClientMockRecorder) UntagResource(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UntagResource", reflect.TypeOf((*MockClient)(nil).UntagResource), input)
}
