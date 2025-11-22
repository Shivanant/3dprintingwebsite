import { Suspense } from "react";
import LoginForm from "./LoginForm";

export default function LoginPage() {
  return (
    <Suspense fallback={<div className="max-w-md mx-auto">Loading...</div>}>
      <LoginForm />
    </Suspense>
  );
}
