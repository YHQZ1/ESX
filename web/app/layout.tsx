import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "ESX — Escrow Stock Exchange",
  description:
    "The exchange infrastructure trusted by brokers, institutions, and market operators. Order matching, clearing, and settlement — built for the demands of modern securities markets.",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
