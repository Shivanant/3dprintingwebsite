import "./globals.css";
import Link from "next/link";
import HeaderNav from "@/components/HeaderNav";

export const metadata = { title: "3DPrint Hub", description: "Sell & print your ideas" };

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body className="min-h-screen bg-slate-50 text-gray-900 flex flex-col">
        <HeaderNav />

        <main className="container mx-auto p-6 flex-1 w-full">{children}</main>

        <footer className="border-t bg-white mt-12 py-8 text-sm text-gray-500">
          <div className="container mx-auto flex flex-wrap items-center gap-6">
            <Link href="/help" className="hover:text-gray-800 transition-colors">
              FAQ
            </Link>
            <Link href="/help#shipping" className="hover:text-gray-800 transition-colors">
              Shipping & returns
            </Link>
            <Link href="/help#support" className="hover:text-gray-800 transition-colors">
              Contact support
            </Link>
            <span className="ml-auto text-xs text-gray-400">
              © {new Date().getFullYear()} 3DPrint Hub
            </span>
          </div>
        </footer>
      </body>
    </html>
  );
}
