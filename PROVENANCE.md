# Bootstrap provenance

Cueia's Go and TypeScript source, database migrations, North Pier invitation, event, registration form, preferences, and browser expectations were authored for this repository. Mara, Pier Hall and the workshop are fictional. The expected results were written independently of the seed loader. No external artwork was used.

The PDF was typeset using the built-in Helvetica face with ReportLab 4.4.3. Its two AcroForm text fields were checked by opening, filling, saving and reopening a copy with pypdf 6.1.3. Both tools are development utilities. [ReportLab](https://www.reportlab.com/) · [pypdf](https://pypdf.readthedocs.io/)

The backend uses pgx and the PostgreSQL wire protocol. The build images distribute Go and Alpine Linux, and Compose runs the official PostgreSQL image. The resolved Go dependency versions and checksums are in backend/go.mod and backend/go.sum. [pgx](https://github.com/jackc/pgx) · [Go](https://go.dev/) · [Alpine](https://alpinelinux.org/) · [PostgreSQL](https://www.postgresql.org/)

The Bun workspace uses Turbo, TypeScript, React, React DOM, Adobe React Spectrum 2, and openapi-typescript. The manifests identify direct dependencies. bun.lock records resolved versions, including transitive packages. [Bun](https://bun.com/) · [Turbo](https://turborepo.com/) · [TypeScript](https://www.typescriptlang.org/) · [React](https://react.dev/) · [Spectrum](https://react-spectrum.adobe.com/) · [openapi-typescript](https://openapi-ts.dev/)

The API contract uses OpenAPI. The planned identity and infrastructure integrations use Clerk session JWTs and the Azure Resource Manager API. [OpenAPI](https://spec.openapis.org/) · [Clerk](https://clerk.com/docs) · [Azure Resource Manager](https://learn.microsoft.com/azure/azure-resource-manager/)

The Azure adapter uses Microsoft's Go SDK, Go's SSH implementation for private file access. [Azure SDK](https://github.com/Azure/azure-sdk-for-go) · [Go SSH](https://pkg.go.dev/golang.org/x/crypto/ssh)
