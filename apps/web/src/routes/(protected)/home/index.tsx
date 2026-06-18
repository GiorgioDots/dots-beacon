import { Button } from "#/components/ui/button";
import keycloak from "#/lib/auth/keycloak";
import { createFileRoute, Link } from "@tanstack/react-router";

export const Route = createFileRoute("/(protected)/home/")({
  component: RouteComponent,
  beforeLoad: async () => {
    const res = await fetch("http://localhost:8080/sites", {
      headers: {
        Authorization: `Bearer ${keycloak.token}`,
      },
    });
    console.log(await res.json());
  },
});

function RouteComponent() {
  return (
    <>
      <Button
        onClick={() => keycloak.logout({ redirectUri: `${location.origin}` })}
      >
        logout
      </Button>
      <Link to="/">Index</Link>
    </>
  );
}
