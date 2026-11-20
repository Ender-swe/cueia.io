# Bootstrap provenance

Cueia's Go and TypeScript source, database migrations, North Pier invitation, event, registration form, preferences, and browser expectations were authored for this repository. Mara, Pier Hall and the workshop are fictional. The expected results were written independently of the seed loader. No external artwork was used.

The PDF was typeset using the built-in Helvetica face with ReportLab 4.4.3. Its two AcroForm text fields were checked by opening, filling, saving and reopening a copy with pypdf 6.1.3. Both tools are development utilities. [ReportLab](https://www.reportlab.com/) · [pypdf](https://pypdf.readthedocs.io/)

The backend uses pgx and the PostgreSQL wire protocol. The build images distribute Go and Alpine Linux, and Compose runs the official PostgreSQL image. The resolved Go dependency versions and checksums are in apps/server/go.mod and apps/server/go.sum. [pgx](https://github.com/jackc/pgx) · [Go](https://go.dev/) · [Alpine](https://alpinelinux.org/) · [PostgreSQL](https://www.postgresql.org/)

The Bun workspace uses Turbo, TypeScript, React, React DOM, Adobe React Spectrum 2, and openapi-typescript. The manifests identify direct dependencies. bun.lock records resolved versions, including transitive packages. [Bun](https://bun.com/) · [Turbo](https://turborepo.com/) · [TypeScript](https://www.typescriptlang.org/) · [React](https://react.dev/) · [Spectrum](https://react-spectrum.adobe.com/) · [openapi-typescript](https://openapi-ts.dev/)

The API contract uses OpenAPI. The planned identity and infrastructure integrations use Clerk session JWTs and the Azure Resource Manager API. [OpenAPI](https://spec.openapis.org/) · [Clerk](https://clerk.com/docs) · [Azure Resource Manager](https://learn.microsoft.com/azure/azure-resource-manager/)

The Azure adapter uses Microsoft's Go SDK, Go's SSH implementation for private file access. [Azure SDK](https://github.com/Azure/azure-sdk-for-go) · [Go SSH](https://pkg.go.dev/golang.org/x/crypto/ssh)

The dashboard uses Vite, React Router and Clerk's React SDK. [Vite](https://vite.dev/) · [React Router](https://reactrouter.com/) · [Clerk React SDK](https://github.com/clerk/javascript)

The workspace controls use Adobe Spectrum icons. Tailwind CSS supplies layout utilities, and Motion for React handles panel fades. The uncodixfy project skill was supplied by the project owner.

The Cueia profile image is an original image of a folded terracotta paper loop generated for this dashboard. The rail mark was drawn in CSS. The sample conversation and its illustrative reminder were written for Cueia using the fictional North Pier invitation and event. The dashboard bundles DM Sans from Fontsource under the SIL Open Font License. [DM Sans](https://github.com/googlefonts/dm-fonts)
