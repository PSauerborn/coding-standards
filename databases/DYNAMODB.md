---
title: DynamoDB Code Standards
description: Standards for DynamoDB single-table design and access patterns.
parent: GENERAL.md
scope: '*'
topics:
- dynamodb
- single-table-design
- nosql
---

# 1. Data Modeling

`[DDB-001]` **MUST**: All DynamoDB tables must have `PK` and `SK` as partition and sort key respectively.

`[DDB-002]` **MUST**: All global secondary indexes must have `GSI{N}PK` and `GSI{N}SK` as partition and sort key respectively, where `{N}` is an integer referring to the index number.

`[DDB-003]` **MUST**: All data stored in a key field must also be stored as a separate attribute in the object. For instance, if an item has partition key `USER#{userId}` then the item must also have a `userId` attribute. This ensures keys can be reconstructed from the non-key item attributes.

`[DDB-004]` **MUST**: All DynamoDB tables must implement a single-table design pattern. All keys must follow the structure `TYPE#{id}` where `TYPE` is the type of the item and `id` is the unique identifier of the item. Keys can have multiple parts, for instance `USER#{userId}` or `USER#{userId}#EMAIL#{email}`. This ensures that it is clear what type of entity is stored in the item from the structure of the key.

`[DDB-005]` **MUST**: Table scan operations are slow and must be avoided. Use global secondary indexes instead.

`[DDB-006]` **MUST**: Persistence layers must use UTC time when storing timestamps.

`[DDB-007]` **MUST**: Persistence layers must use transactions where possible to ensure data consistency. This helps to prevent data loss and partial updates.

`[DDB-008]` **SHOULD**: All DynamoDB items should have a `createdAt` and `updatedAt` attribute to track when the item was created and last updated.

`[DDB-009]` **SHOULD**: The number of items for any given global secondary index should be kept relatively small to avoid large partitions. Prefer duplicating data over several items with different keys/access patterns.

`[DDB-010]` **SHOULD**: Persistence layers should use `BatchGetItem` and `BatchWriteItem` operations to reduce the number of individual requests to DynamoDB.

## Example 1

The following example illustrates how DynamoDB items should be structured and keyed:

```json
// GOOD: use PK and SK for primary key
// GOOD: use GSI1PK and GSI1SK for global secondary index
// GOOD: use userId, email and organizationId (data that appears in keys) is also stored as separate attributes
// GOOD: createdAt and updatedAt attributes exist with UTC timestamps.
{
    "PK": "USER#123", 
    "SK": "PROFILE",
    "GSI1PK": "EMAIL#test@example.com", 
    "GSI1SK": "EMAIL#test@example.com",
    "GSI2PK": "ORGANIZATION_ID#org123", 
    "GSI2SK": "CREATED_TS#2025-12-28T15:35:25Z",
    "userId": "123",
    "organizationId": "org123",
    "email": "test@example.com",
    "firstName": "Pascal",
    "lastName": "Sauerborn",
    "createdAt": "2025-12-28T15:35:25Z",
    "updatedAt": "2025-12-28T15:35:25Z"
}
```