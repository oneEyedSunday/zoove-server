# Migration `20200929081923-another-banger`

This migration has been generated by Ayomide Onigbinde <Onigbindeayomide@gmail.com> at 9/29/2020, 9:19:23 AM.
You can check out the [state of the schema](./schema.prisma) after the migration.

## Database Steps

```sql
ALTER TABLE "public"."User" DROP COLUMN "token"
```

## Changes

```diff
diff --git schema.prisma schema.prisma
migration 20200929071101-another-banger..20200929081923-another-banger
--- datamodel.dml
+++ datamodel.dml
@@ -1,6 +1,6 @@
 datasource postgresql {
-  url = "***"
+  url = "***"
   provider = "postgresql"
 }
 generator db {
@@ -21,6 +21,5 @@
   email     String   @unique
   username  String   @unique
   platform  String
   avatar    String
-  token     String
 }
```

