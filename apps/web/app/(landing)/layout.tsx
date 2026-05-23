import { cookies } from "next/headers";
import { Instrument_Serif } from "next/font/google";
import { LocaleProvider } from "@/features/landing/i18n";
import type { Locale } from "@/features/landing/i18n";

const instrumentSerif = Instrument_Serif({
  subsets: ["latin"],
  weight: "400",
  variable: "--font-serif",
});

const jsonLd = {
  "@context": "https://schema.org",
  "@graph": [
    {
      "@type": "Organization",
      name: "Cimeria",
      url: "https://app.cimeria.online",
    },
    {
      "@type": "SoftwareApplication",
      name: "Cimeria",
      applicationCategory: "BusinessApplication",
      operatingSystem: "Web",
      description:
        "AI-powered sales automation platform.",
      offers: {
        "@type": "Offer",
        price: "0",
        priceCurrency: "USD",
      },
    },
  ],
};

async function getInitialLocale(): Promise<Locale> {
  // Keep the public Cimeria landing in English until localized copy is maintained.
  const cookieStore = await cookies();
  const stored = cookieStore.get("cimeria-locale")?.value;
  if (stored === "en") return stored;

  return "en";
}

export default async function LandingLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const initialLocale = await getInitialLocale();

  return (
    <>
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }}
      />
      <div className={`${instrumentSerif.variable} h-full overflow-x-hidden overflow-y-auto bg-white`}>
        <LocaleProvider initialLocale={initialLocale}>{children}</LocaleProvider>
      </div>
    </>
  );
}
