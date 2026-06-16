import { Button } from "#/components/ui/button";
import keycloak from "#/lib/auth/keycloak";
import { createFileRoute } from "@tanstack/react-router";
import { useEffect } from "react";

export const Route = createFileRoute("/")({ component: Home });

function Home() {
  useEffect(() => {
    async function test() {
      try {
        // const authenticated = await keycloak.init();
        // if (authenticated) {
        //   console.log("User is authenticated");
        // } else {
        //   console.log("User is not authenticated");
        // }
        console.log(keycloak.token);
      } catch (error) {
        console.error("Failed to initialize adapter:", error);
      }
    }
    test();
  });
  return (
    <>
      <Button onClick={() => keycloak.logout()}>logout</Button>
    </>
  );
}
