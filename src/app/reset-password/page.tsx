import { Suspense } from "react";
import ResetPasswordForm from "./ResetPasswordForm";

export default function ResetPasswordPage() {
  return (
    <Suspense fallback={<div className="max-w-md mx-auto">Loading...</div>}>
      <ResetPasswordForm />
    </Suspense>
  );
}
