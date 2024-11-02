CREATE TABLE "public"."preference" (
    "user_id" int8 NOT NULL,
    "notify_on_update" bool NOT NULL DEFAULT false,
    PRIMARY KEY ("user_id")
);