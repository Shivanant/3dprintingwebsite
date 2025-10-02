import Link from "next/link";

export default function Home() {
  return (
    <div className="space-y-10">
      <section className="rounded-2xl border bg-white p-10">
        <div className="grid md:grid-cols-2 gap-8 items-center">
          <div>
            <h1 className="text-4xl font-extrabold">3D Printing Products & Custom Jobs</h1>
            <p className="mt-3 text-gray-600">
              Upload STL/G-code for instant preview & pricing. Or turn a photo into a lithophane.
            </p>
            <div className="mt-6 flex gap-3">
              <Link className="px-4 py-2 rounded bg-black text-white" href="/custom-print">Start Custom Print</Link>
              <Link className="px-4 py-2 rounded border" href="/lithophane">Create Lithophane</Link>
            </div>
          </div>
          <div className="h-56 rounded-xl bg-gray-100 flex items-center justify-center">
            <span className="text-gray-400">Hero Preview</span>
          </div>
        </div>
      </section>

      <section className="grid md:grid-cols-3 gap-6">
        {[
          { title: "Shop Products", href: "/shop" },
          { title: "Custom 3D Print", href: "/custom-print" },
          { title: "Lithophane", href: "/lithophane" },
        ].map((c) => (
          <Link key={c.title} href={c.href}
            className="rounded-xl border bg-white p-6 hover:shadow">
            <div className="h-32 bg-gray-100 rounded mb-4" />
            <h3 className="font-semibold">{c.title}</h3>
            <p className="text-sm text-gray-600">Go to {c.title}</p>
          </Link>
        ))}
      </section>
    </div>
  );
}
