<div align="center">

# go-fiber-template

![img](https://cdn.discordapp.com/attachments/372372440334073859/1216723412244893738/go_1.gif?ex=66016cfb&is=65eef7fb&hm=76c9538616341d981031c54ade56cbeee7eadc96c71d64bbaa6f85cc60436b40&)

</div>

## Install golang package

```bash
go mod tidy
```

# Set Environment `.env`
```
SUPABASE_URL= <your Supabase project URL>
SUPABASE_KEY= <your Supabase API key>
PORT=1818

JWT_SECRET_KEY=Test
JWT_REFESH_SECRET_KEY=Test
```

# Stock movements API

- `GET /api/v1/stock-movements` lists stock movements.
- `POST /api/v1/stock-movements` creates one. Provide `product_id`, `type` (`IN` or `OUT`), `quantity` (> 0), `previous_stock`, and `current_stock`; `reason` is optional. The database generates `id` and `created_at`.

# Products API

- `GET /api/v1/products` lists products.
- `GET /api/v1/products/:id` gets one product by UUID.
- `POST /api/v1/products` creates a product. Required fields: `name`, `sku`, and `price`. `category_id`, `stock`, `min_stock`, and `unit` are optional; database defaults apply when omitted.
- `PATCH /api/v1/products/:id` updates supplied product fields.
- `DELETE /api/v1/products/:id` deletes a product. Products referenced by stock movements cannot be deleted.

# Categories API

- `GET /api/v1/categories` lists categories.
- `GET /api/v1/categories/:id` gets one category by UUID.
- `POST /api/v1/categories` creates a category. `name` is required; `description` is optional and may be `null`. The database generates `id` and `created_at`.
- `PATCH /api/v1/categories/:id` updates `name` and/or `description`; set `description` to `null` to clear it.
- `DELETE /api/v1/categories/:id` deletes a category. Categories referenced by products cannot be deleted.

# Start APP

```sh
go run . || go run main.go
```

## Air

```sh
go install github.com/cosmtrek/air@latest
```

```sh
air
```

# Podman

```
podman build -t fiber-test .
```

```
podman run --rm -it -p 3000:3000 fiber-test
```
