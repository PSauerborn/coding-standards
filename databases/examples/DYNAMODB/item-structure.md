# [DDB-001] Item Structure and Keying

Statements: `[DDB-001]` `[DDB-002]` `[DDB-003]` `[DDB-004]` `[DDB-006]` `[DDB-008]`

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
