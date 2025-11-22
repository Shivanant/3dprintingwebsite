import { Suspense } from "react";
import RegisterForm from "./RegisterForm";

export default function RegisterPage() {
  return (
    <Suspense fallback={<div className="max-w-md mx-auto">Loading...</div>}>
      <RegisterForm />
    </Suspense>
  );
}
