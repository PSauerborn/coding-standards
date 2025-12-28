# DynamoDB Code Standards

# 1. Meta Rules

You are a Senior Software Engineer acting as an autonomous coding agent.
1.  **Strict Adherence**: You MUST follow all **MUST** rules below.
2.  **Pattern Matching**: When writing code, check the "Example" sections. If you are tempted to write code that looks like a "BAD" example, STOP and refactor to match the "GOOD" example.
3.  **Explanation**: If you deviate from a **SHOULD** rule, you must explicitly state why in your reasoning trace.

If a user request contradicts a **SHOULD** statement, follow the user request. If it contradicts a **MUST** statement, ask for confirmation.

# 2. Data Modeling

**MUST**: All DynamoDB tables must have `PK` and `SK` as partition and sort key respectively.

**MUST**: All global secondary indexes must have `GSI{N}PK` and `GSI{N}SK` as partition and sort key respectively, where `{N}` is an integer refering to the index number.

**MUST**: All data stored in a key field must also be stored as a separate attribute in the object. For instance, if an item has partition key `USER#{userId}` then the item must also have a `userId` attribute. This ensures keys can be reconstructed from the non-key item attributes.

**MUST**: All DynamoDB tables must implement a single-table design pattern. All keys must follow the structure `TYPE#{id}` where `TYPE` is the type of the item and `id` is the unique identifier of the item. Keys can have multiple parts, for instance `USER#{userId}` or `USER#{userId}#EMAIL#{email}`. This ensures that it is clear what type of entity is stored in the item from the structure of the key.

**MUST**: Table scan operations are slow and must be avoided. Use global secondary indexes instead.

**MUST**: Persistence layers must use UTC time when storing timestamps.

**SHOULD**: All DynamoDBN items should have a `createdAt` and `updatedAt` attribute to track when the item was created and last updated.

**SHOULD**: The number of items for any given global secondary index should be kept relatively small to avoid large partitions. Prefer duplicating data over several items with different keys/access patterns.

**MUST**: Persistence layers must use transactions where possible to ensure data consistency. This helpls to prevent data loss and partial updates.

**SHOULD**: Persistence layers should use `BatchGetItem` and `BatchWriteItem` operations to reduce the number of individual requests to DynamoDB.

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