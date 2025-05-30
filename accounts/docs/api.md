# Accounts Service API Documentation

The Accounts Service provides a GraphQL API for managing accounts, vendors, products, and orders.

## GraphQL Endpoint

The GraphQL API is available at:

```
http://localhost:4000/graphql
```

A GraphQL Playground is available at:

```
http://localhost:4000/playground
```

## Authentication

Authentication is handled via JWT tokens. To authenticate, include the JWT token in the Authorization header:

```
Authorization: Bearer <token>
```

## Types

### Account

Represents a user account in the system.

```graphql
type Account {
  id: ID!
  username: String!
  email: String!
  firstName: String
  lastName: String
  isActive: Boolean!
  isAdmin: Boolean!
  createdAt: String!
  updatedAt: String!
  orders: [Order!]
}
```

### Vendor

Represents a vendor in the system.

```graphql
type Vendor {
  id: ID!
  name: String!
  description: String
  contactEmail: String
  contactPhone: String
  address: String
  isActive: Boolean!
  createdAt: String!
  updatedAt: String!
  products: [Product!]
}
```

### Product

Represents a product in the system.

```graphql
type Product {
  id: ID!
  vendorId: ID!
  name: String!
  description: String
  price: Float!
  isAvailable: Boolean!
  createdAt: String!
  updatedAt: String!
  vendor: Vendor
}
```

### Order

Represents an order in the system.

```graphql
type Order {
  id: ID!
  accountId: ID!
  status: OrderStatus!
  totalAmount: Float!
  createdAt: String!
  updatedAt: String!
  account: Account
  items: [OrderItem!]
}
```

### OrderItem

Represents an item in an order.

```graphql
type OrderItem {
  id: ID!
  orderId: ID!
  productId: ID!
  quantity: Int!
  unitPrice: Float!
  totalPrice: Float!
  createdAt: String!
  product: Product
}
```

### OrderStatus

Represents the status of an order.

```graphql
enum OrderStatus {
  PENDING
  PAID
  SHIPPED
  DELIVERED
  CANCELLED
}
```

## Queries

### Account Queries

```graphql
# Get an account by ID
query GetAccount($id: ID!) {
  account(id: $id) {
    id
    username
    email
    firstName
    lastName
    isActive
    isAdmin
    createdAt
    updatedAt
  }
}

# List accounts with pagination
query ListAccounts($offset: Int, $limit: Int) {
  accounts(offset: $offset, limit: $limit) {
    id
    username
    email
    firstName
    lastName
    isActive
    isAdmin
    createdAt
    updatedAt
  }
  accountsCount
}
```

### Vendor Queries

```graphql
# Get a vendor by ID
query GetVendor($id: ID!) {
  vendor(id: $id) {
    id
    name
    description
    contactEmail
    contactPhone
    address
    isActive
    createdAt
    updatedAt
  }
}

# List vendors with pagination
query ListVendors($offset: Int, $limit: Int) {
  vendors(offset: $offset, limit: $limit) {
    id
    name
    description
    contactEmail
    contactPhone
    address
    isActive
    createdAt
    updatedAt
  }
  vendorsCount
}

# Search vendors by name
query SearchVendors($query: String!, $offset: Int, $limit: Int) {
  searchVendors(query: $query, offset: $offset, limit: $limit) {
    id
    name
    description
    contactEmail
    contactPhone
    address
    isActive
    createdAt
    updatedAt
  }
}
```

### Product Queries

```graphql
# Get a product by ID
query GetProduct($id: ID!) {
  product(id: $id) {
    id
    vendorId
    name
    description
    price
    isAvailable
    createdAt
    updatedAt
    vendor {
      id
      name
    }
  }
}

# List products with pagination
query ListProducts($offset: Int, $limit: Int) {
  products(offset: $offset, limit: $limit) {
    id
    vendorId
    name
    description
    price
    isAvailable
    createdAt
    updatedAt
    vendor {
      id
      name
    }
  }
  productsCount
}

# List products by vendor
query ListProductsByVendor($vendorId: ID!, $offset: Int, $limit: Int) {
  productsByVendor(vendorId: $vendorId, offset: $offset, limit: $limit) {
    id
    vendorId
    name
    description
    price
    isAvailable
    createdAt
    updatedAt
  }
}

# Search products by name
query SearchProducts($query: String!, $offset: Int, $limit: Int) {
  searchProducts(query: $query, offset: $offset, limit: $limit) {
    id
    vendorId
    name
    description
    price
    isAvailable
    createdAt
    updatedAt
    vendor {
      id
      name
    }
  }
}
```

### Order Queries

```graphql
# Get an order by ID
query GetOrder($id: ID!) {
  order(id: $id) {
    id
    accountId
    status
    totalAmount
    createdAt
    updatedAt
    account {
      id
      username
    }
    items {
      id
      productId
      quantity
      unitPrice
      totalPrice
      product {
        id
        name
      }
    }
  }
}

# List orders with pagination
query ListOrders($offset: Int, $limit: Int) {
  orders(offset: $offset, limit: $limit) {
    id
    accountId
    status
    totalAmount
    createdAt
    updatedAt
  }
  ordersCount
}

# List orders by account
query ListOrdersByAccount($accountId: ID!, $offset: Int, $limit: Int) {
  ordersByAccount(accountId: $accountId, offset: $offset, limit: $limit) {
    id
    status
    totalAmount
    createdAt
    updatedAt
    items {
      id
      productId
      quantity
      unitPrice
      totalPrice
      product {
        id
        name
      }
    }
  }
}

# List orders by status
query ListOrdersByStatus($status: OrderStatus!, $offset: Int, $limit: Int) {
  ordersByStatus(status: $status, offset: $offset, limit: $limit) {
    id
    accountId
    status
    totalAmount
    createdAt
    updatedAt
    account {
      id
      username
    }
  }
}
```

## Mutations

### Account Mutations

```graphql
# Create a new account
mutation CreateAccount($input: CreateAccountInput!) {
  createAccount(input: $input) {
    id
    username
    email
    firstName
    lastName
    isActive
    isAdmin
    createdAt
    updatedAt
  }
}

# Update an existing account
mutation UpdateAccount($id: ID!, $input: UpdateAccountInput!) {
  updateAccount(id: $id, input: $input) {
    id
    username
    email
    firstName
    lastName
    isActive
    isAdmin
    createdAt
    updatedAt
  }
}

# Delete an account
mutation DeleteAccount($id: ID!) {
  deleteAccount(id: $id)
}
```

### Vendor Mutations

```graphql
# Create a new vendor
mutation CreateVendor($input: CreateVendorInput!) {
  createVendor(input: $input) {
    id
    name
    description
    contactEmail
    contactPhone
    address
    isActive
    createdAt
    updatedAt
  }
}

# Update an existing vendor
mutation UpdateVendor($id: ID!, $input: UpdateVendorInput!) {
  updateVendor(id: $id, input: $input) {
    id
    name
    description
    contactEmail
    contactPhone
    address
    isActive
    createdAt
    updatedAt
  }
}

# Delete a vendor
mutation DeleteVendor($id: ID!) {
  deleteVendor(id: $id)
}
```

### Product Mutations

```graphql
# Create a new product
mutation CreateProduct($input: CreateProductInput!) {
  createProduct(input: $input) {
    id
    vendorId
    name
    description
    price
    isAvailable
    createdAt
    updatedAt
  }
}

# Update an existing product
mutation UpdateProduct($id: ID!, $input: UpdateProductInput!) {
  updateProduct(id: $id, input: $input) {
    id
    vendorId
    name
    description
    price
    isAvailable
    createdAt
    updatedAt
  }
}

# Delete a product
mutation DeleteProduct($id: ID!) {
  deleteProduct(id: $id)
}
```

### Order Mutations

```graphql
# Create a new order
mutation CreateOrder($input: CreateOrderInput!) {
  createOrder(input: $input) {
    id
    accountId
    status
    totalAmount
    createdAt
    updatedAt
    items {
      id
      productId
      quantity
      unitPrice
      totalPrice
    }
  }
}

# Update an order status
mutation UpdateOrderStatus($id: ID!, $status: OrderStatus!) {
  updateOrderStatus(id: $id, status: $status) {
    id
    status
    updatedAt
  }
}

# Delete an order
mutation DeleteOrder($id: ID!) {
  deleteOrder(id: $id)
}
```
