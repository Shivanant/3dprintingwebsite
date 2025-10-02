import "./globals.css";
import Link from "next/link";

export const metadata = { title: "3DPrint Hub", description: "Sell & print your ideas" };

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body className="min-h-screen bg-gray-50 text-gray-900">
        <header className="border-b bg-white">
          <nav className="container mx-auto flex items-center gap-6 p-4">
            <Link href="/" className="font-bold">3DPrint Hub</Link>
            <div className="flex items-center gap-4 text-sm">
              <Link href="/shop">Shop</Link>
              <Link href="/custom-print">Custom Print</Link>
              <Link href="/lithophane">Lithophane</Link>
              <Link href="/account">Account</Link>
              <Link href="/help">Help</Link>
            </div>
            <div className="ml-auto">
              <Link href="/cart">🛒 Cart</Link>
            </div>
          </nav>
        </header>

        <main className="container mx-auto p-6">{children}</main>

        <footer className="border-t mt-12 py-8 text-sm text-gray-500">
          <div className="container mx-auto flex gap-6">
            <Link href="/help">FAQ</Link>
            <span>•</span>
            <span>Policies</span>
            <span>•</span>
            <span>Contact</span>
          </div>
        </footer>
      </body>
    </html>
  );
}
