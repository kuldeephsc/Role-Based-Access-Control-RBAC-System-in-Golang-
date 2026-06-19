import "./globals.css";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "RBAC Platform",
  description: "Role-based access control dashboard",
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
