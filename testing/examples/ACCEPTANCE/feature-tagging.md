# [ACPT-007] Feature File Tagging

Statements: `[ACPT-006]` `[ACPT-007]` `[ACPT-008]` `[ACPT-009]` `[ACPT-010]`

Feature files should cover a single functional domain, and scenarios should be tagged with the component they exercise and the source requirement they verify. The following is an example of a well-tagged feature file:

```gherkin
# File: acceptance/features/orders.feature

# GOOD: optional domain tag applied once at the Feature level
@orders
Feature: Order placement
    Customers can place orders and receive confirmation.

    # GOOD: component tag (@api) plus source requirement tag (@spec-004)
    @api @spec-004
    Scenario: A customer places an order for an in-stock item
        Given the item "Standing Desk" is in stock
        When the customer places an order for "Standing Desk"
        Then the order is confirmed
        And the customer receives an order confirmation email

    # GOOD: requirement tags may reference tickets or user stories instead of specs
    @ui @JIRA-451
    Scenario: An out-of-stock item cannot be ordered
        Given the item "Standing Desk" is out of stock
        When the customer views the product page for "Standing Desk"
        Then the option to place an order is not available
```

Conversely, the following feature file mixes domains, provides no traceability, and is written in technical language that non-technical stakeholders cannot read:

```gherkin
# File: acceptance/features/misc.feature

# BAD: file is not scoped to a single functional domain
# BAD: no plain-language description of the value the feature provides
Feature: Misc tests

    # BAD: no component tag and no source requirement tag —
    # the scenario cannot be traced back to a spec, ticket, or user story
    Scenario: OrderService.PlaceOrder inserts a row
        # BAD: steps reference implementation details (tables, functions,
        # serialized payloads) instead of observable behavior
        Given a row exists in the "inventory" table with stock_count > 0
        When OrderService.PlaceOrder is invoked with a valid OrderRequest struct
        Then a new row is inserted into the "orders" table with status_id = 2
        And the confirmation goroutine enqueues a message on the "emails" channel

    # BAD: unrelated domain (payments) in the same feature file
    Scenario: A payment is taken for a confirmed order
        Given an order for "Standing Desk" is confirmed
        When the payment is processed
        Then the customer is charged for the order
```
