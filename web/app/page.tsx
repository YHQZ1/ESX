/* eslint-disable @next/next/no-img-element */
"use client";

import { useState, useEffect } from "react";

/* ─── Data ─────────────────────────────────────────────────── */

const TICKERS = [
  { sym: "RELIANCE", price: "2,847.50", chg: "+1.24%", up: true },
  { sym: "TCS", price: "3,921.00", chg: "+0.87%", up: true },
  { sym: "INFY", price: "1,582.75", chg: "-0.43%", up: false },
  { sym: "HDFCBANK", price: "1,674.20", chg: "+2.01%", up: true },
  { sym: "BAJFIN", price: "7,203.85", chg: "+3.14%", up: true },
  { sym: "WIPRO", price: "456.30", chg: "-0.12%", up: false },
  { sym: "ICICIBNK", price: "1,089.40", chg: "+0.66%", up: true },
  { sym: "TATAMOT", price: "878.15", chg: "-1.02%", up: false },
  { sym: "SUNPHRM", price: "1,341.60", chg: "+0.29%", up: true },
  { sym: "MARUTI", price: "10,456.00", chg: "+1.78%", up: true },
  { sym: "LT", price: "3,672.90", chg: "+0.55%", up: true },
  { sym: "AXISBNK", price: "1,123.75", chg: "-0.34%", up: false },
];

const INDICES = [
  {
    name: "ESX Composite",
    value: "24,812.44",
    chg: "+183.20",
    pct: "+0.74%",
    up: true,
  },
  {
    name: "ESX 50",
    value: "7,341.20",
    chg: "-22.60",
    pct: "-0.31%",
    up: false,
  },
  {
    name: "ESX MidCap",
    value: "11,094.80",
    chg: "+67.40",
    pct: "+0.61%",
    up: true,
  },
  {
    name: "ESX SmallCap",
    value: "4,218.35",
    chg: "+31.15",
    pct: "+0.74%",
    up: true,
  },
];

const SERVICES = [
  {
    n: "01",
    title: "Order Matching Engine",
    body: "Price-time priority matching across all listed securities. Every order type — market, limit, stop, IOC — processed with strict FIFO discipline and full circuit-breaker protection at the security level.",
    metric: "< 2µs",
    mLabel: "match latency",
  },
  {
    n: "02",
    title: "Central Counterparty Clearing",
    body: "ESX novates into every executed trade as central counterparty, guaranteeing settlement to both sides regardless of the other's default. The same model used by DTCC, NSCCL, and LCH Clearnet.",
    metric: "100%",
    mLabel: "trade guarantee",
  },
  {
    n: "03",
    title: "Atomic DvP Settlement",
    body: "Delivery versus Payment enforced through a single atomic transaction. Cash and securities transfer simultaneously — or neither does. No partial settlement. No open exposure between execution and finality.",
    metric: "T+0",
    mLabel: "settlement",
  },
  {
    n: "04",
    title: "Pre-Trade Risk Controls",
    body: "Every order is validated against real-time collateral before it touches the book. Cash locked for buys, securities locked for sells. Undercollateralised orders are rejected at the gateway — never matched.",
    metric: "Zero",
    mLabel: "bad fills",
  },
  {
    n: "05",
    title: "Market Data Services",
    body: "Live trade feeds, OHLCV candlesticks, order book depth snapshots, and last-traded prices — delivered over WebSocket to connected trading systems, data vendors, and downstream platforms.",
    metric: "< 5ms",
    mLabel: "feed latency",
  },
  {
    n: "06",
    title: "FIX Protocol Connectivity",
    body: "Native FIX 4.2 implementation. Any institutional broker, algorithmic trading system, or order management platform connects without protocol translation — the standard that has carried global volume for 30 years.",
    metric: "FIX 4.2",
    mLabel: "native protocol",
  },
];

const LIFECYCLE = [
  {
    step: "1",
    title: "Order Submission",
    desc: "Brokers and participants submit orders via FIX 4.2 or the REST API. The gateway parses, authenticates, and validates every message.",
  },
  {
    step: "2",
    title: "Risk Validation",
    desc: "The risk engine checks collateral in real time. Required cash or securities are locked before the order proceeds. Deficient orders are rejected immediately.",
  },
  {
    step: "3",
    title: "Matching",
    desc: "Orders enter the live order book. Price-time priority fires against resting orders — full fill, partial fill, or the order rests until a counterpart arrives.",
  },
  {
    step: "4",
    title: "Clearing",
    desc: "ESX novates as central counterparty to both sides. The original bilateral exposure is extinguished and replaced with two new obligations — to and from ESX.",
  },
  {
    step: "5",
    title: "Settlement",
    desc: "A single atomic database transaction debits and credits all four legs. Cash moves. Securities move. Simultaneously. If any leg fails, all four roll back.",
  },
  {
    step: "6",
    title: "Ledger Recording",
    desc: "Double-entry journal entries are written for every movement across the cash and securities ledgers. Debits always equal credits. The record is permanent.",
  },
];

const MEMBERS = [
  {
    title: "Trading Members",
    sub: "Brokers & Dealers",
    items: [
      "Direct market access over FIX 4.2",
      "Real-time position and balance APIs",
      "Pre-trade margin and collateral controls",
      "Trade confirmation and settlement reports",
    ],
  },
  {
    title: "Clearing Members",
    sub: "Custodians & Banks",
    items: [
      "Guaranteed settlement through ESX CCP",
      "Netting across positions and instruments",
      "Real-time margin call and mark-to-market",
      "Full audit trail and regulatory reporting",
    ],
  },
  {
    title: "Algorithmic Participants",
    sub: "HFT & Quant Firms",
    items: [
      "Co-location and proximity hosting",
      "Ultra-low latency order path",
      "WebSocket market data at source",
      "Order book depth and time & sales",
    ],
  },
  {
    title: "Listed Companies",
    sub: "Issuers & Corporates",
    items: [
      "Primary listing and IPO infrastructure",
      "Investor registry and share register",
      "Rights issue and offer-for-sale mechanics",
      "Real-time shareholder visibility",
    ],
  },
];

/* ─── Shared styles ─────────────────────────────────────────── */

const mono: React.CSSProperties = {
  fontFamily: "'Geist Mono', monospace",
};

const sans: React.CSSProperties = {
  fontFamily: "'Geist', sans-serif",
};

const title: React.CSSProperties = {
  fontFamily: "'Geist', sans-serif",
  fontWeight: 800,
  textTransform: "uppercase",
  letterSpacing: "-0.03em",
};

/* ─── Ticker ────────────────────────────────────────────────── */

function Ticker() {
  const items = [...TICKERS, ...TICKERS];
  return (
    <div
      style={{
        background: "var(--navy)",
        overflow: "hidden",
        height: "36px",
        display: "flex",
        alignItems: "center",
      }}
    >
      <div
        style={{
          display: "flex",
          animation: "ticker 45s linear infinite",
          width: "max-content",
        }}
      >
        {items.map((t, i) => (
          <span
            key={i}
            style={{
              display: "inline-flex",
              alignItems: "center",
              gap: "10px",
              padding: "0 28px",
              borderRight: "1px solid rgba(255,255,255,0.1)",
              ...mono,
              fontSize: "11px",
            }}
          >
            <span
              style={{
                color: "rgba(255,255,255,0.5)",
                letterSpacing: "0.06em",
              }}
            >
              {t.sym}
            </span>
            <span style={{ color: "#ffffff", fontWeight: 500 }}>{t.price}</span>
            <span style={{ color: t.up ? "var(--up)" : "var(--down)" }}>
              {t.chg}
            </span>
          </span>
        ))}
      </div>
    </div>
  );
}

/* ─── Nav ───────────────────────────────────────────────────── */

function Nav() {
  const [scrolled, setScrolled] = useState(false);

  useEffect(() => {
    const fn = () => setScrolled(window.scrollY > 10);
    window.addEventListener("scroll", fn);
    return () => window.removeEventListener("scroll", fn);
  }, []);

  return (
    <nav
      style={{
        position: "sticky",
        top: 0,
        zIndex: 100,
        background: scrolled ? "rgba(255,255,255,0.95)" : "#ffffff",
        borderBottom: "1px solid var(--border)",
        backdropFilter: scrolled ? "blur(12px)" : "none",
        transition: "all 0.15s ease",
      }}
    >
      <div
        style={{
          padding: "0 56px",
          height: "64px",
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
        }}
      >
        {/* Logo */}
        <div style={{ display: "flex", alignItems: "center", gap: "14px" }}>
          {/* Logo container using your custom png */}
          <div
            style={{
              width: "32px",
              height: "32px",
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
            }}
          >
            <img
              src="/logo.png"
              alt="ESX Logo"
              style={{
                width: "100%",
                height: "100%",
                objectFit: "contain",
              }}
            />
          </div>

          {/* Brand Text */}
          <div>
            <span
              style={{
                ...title,
                fontSize: "18px",
                color: "var(--navy)",
              }}
            >
              ESX
            </span>
          </div>
        </div>

        {/* Links */}
        <div style={{ display: "flex", alignItems: "center", gap: "36px" }}>
          {["Markets", "Products", "Participants", "Data", "Regulations"].map(
            (l) => (
              <a
                key={l}
                href="#"
                style={{
                  ...sans,
                  fontSize: "12px",
                  fontWeight: 600,
                  textTransform: "uppercase",
                  letterSpacing: "0.05em",
                  color: "var(--text-muted)",
                  textDecoration: "none",
                  transition: "color 0.15s",
                }}
                onMouseEnter={(e) =>
                  (e.currentTarget.style.color = "var(--accent)")
                }
                onMouseLeave={(e) =>
                  (e.currentTarget.style.color = "var(--text-muted)")
                }
              >
                {l}
              </a>
            ),
          )}
        </div>

        <div style={{ display: "flex", gap: "16px", alignItems: "center" }}>
          <a
            href="#"
            style={{
              ...sans,
              fontSize: "12px",
              fontWeight: 600,
              textTransform: "uppercase",
              letterSpacing: "0.05em",
              color: "var(--navy)",
              textDecoration: "none",
              padding: "7px 16px",
            }}
          >
            Sign In
          </a>
          <a
            href="#"
            style={{
              ...sans,
              fontSize: "12px",
              fontWeight: 600,
              textTransform: "uppercase",
              letterSpacing: "0.05em",
              background: "var(--navy)",
              color: "#ffffff",
              textDecoration: "none",
              padding: "10px 24px",
              borderRadius: "0px",
              transition: "background 0.15s",
            }}
            onMouseEnter={(e) =>
              (e.currentTarget.style.background = "var(--accent)")
            }
            onMouseLeave={(e) =>
              (e.currentTarget.style.background = "var(--navy)")
            }
          >
            Apply for Membership
          </a>
        </div>
      </div>
    </nav>
  );
}

/* ─── Index bar ─────────────────────────────────────────────── */

function IndexBar() {
  return (
    <div
      style={{
        background: "var(--bg-deep)",
        borderBottom: "1px solid var(--border)",
        padding: "0 56px",
        display: "flex",
        alignItems: "stretch",
        height: "56px",
        overflow: "hidden",
      }}
    >
      {INDICES.map((idx, i) => (
        <div
          key={idx.name}
          style={{
            display: "flex",
            alignItems: "center",
            gap: "16px",
            paddingRight: "40px",
            marginRight: "40px",
            borderRight:
              i < INDICES.length - 1 ? "1px solid var(--border)" : "none",
          }}
        >
          <div>
            <div
              style={{
                ...sans,
                fontSize: "10px",
                color: "var(--text-muted)",
                fontWeight: 600,
                textTransform: "uppercase",
                letterSpacing: "0.05em",
                marginBottom: "2px",
              }}
            >
              {idx.name}
            </div>
            <div
              style={{ display: "flex", alignItems: "baseline", gap: "10px" }}
            >
              <span
                style={{
                  ...mono,
                  fontSize: "14px",
                  fontWeight: 600,
                  color: "var(--navy)",
                }}
              >
                {idx.value}
              </span>
              <span
                style={{
                  ...mono,
                  fontSize: "11px",
                  color: idx.up ? "var(--up)" : "var(--down)",
                }}
              >
                {idx.chg} ({idx.pct})
              </span>
            </div>
          </div>
        </div>
      ))}
      <div
        style={{
          marginLeft: "auto",
          display: "flex",
          alignItems: "center",
          gap: "10px",
        }}
      >
        <span
          style={{
            width: "8px",
            height: "8px",
            background: "var(--up)",
            borderRadius: "0px",
            animation: "blink 2s ease-in-out infinite",
          }}
        />
        <span
          style={{
            ...sans,
            fontSize: "11px",
            color: "var(--up)",
            fontWeight: 600,
            textTransform: "uppercase",
            letterSpacing: "0.05em",
          }}
        >
          Market Open
        </span>
        <span
          style={{
            ...mono,
            fontSize: "11px",
            color: "var(--text-subtle)",
            marginLeft: "12px",
          }}
        >
          IST · 14:32:41
        </span>
      </div>
    </div>
  );
}

/* ─── Hero ──────────────────────────────────────────────────── */

function Hero() {
  const [time, setTime] = useState("14:32:41.382");

  useEffect(() => {
    const tick = () => {
      const now = new Date();
      const h = String(now.getHours()).padStart(2, "0");
      const m = String(now.getMinutes()).padStart(2, "0");
      const s = String(now.getSeconds()).padStart(2, "0");
      const ms = String(now.getMilliseconds()).padStart(3, "0");
      setTime(`${h}:${m}:${s}.${ms}`);
    };
    const id = setInterval(tick, 80);
    return () => clearInterval(id);
  }, []);

  const asks = [
    { p: "2,848.75", q: "340", v: "973.5K", d: 28 },
    { p: "2,848.50", q: "520", v: "1.48M", d: 43 },
    { p: "2,848.25", q: "760", v: "2.16M", d: 63 },
    { p: "2,848.00", q: "1,100", v: "3.13M", d: 91 },
    { p: "2,847.75", q: "980", v: "2.79M", d: 81 },
  ];
  const bids = [
    { p: "2,847.50", q: "1,200", v: "3.41M", d: 100 },
    { p: "2,847.25", q: "850", v: "2.42M", d: 70 },
    { p: "2,847.00", q: "620", v: "1.76M", d: 51 },
    { p: "2,846.75", q: "440", v: "1.25M", d: 36 },
    { p: "2,846.50", q: "290", v: "825.4K", d: 24 },
  ];
  const trades = [
    { t: "14:32:41", p: "2,847.50", q: "200", side: true },
    { t: "14:32:38", p: "2,847.25", q: "150", side: false },
    { t: "14:32:35", p: "2,847.50", q: "500", side: true },
    { t: "14:32:31", p: "2,847.00", q: "75", side: true },
    { t: "14:32:28", p: "2,846.75", q: "300", side: false },
    { t: "14:32:22", p: "2,847.25", q: "120", side: true },
  ];

  return (
    <section
      style={{
        background: "#ffffff",
        borderBottom: "1px solid var(--border)",
        display: "grid",
        gridTemplateColumns: "1fr 520px", // Increased terminal width to fill horizontal space
        minHeight: "580px", // Reduced from 640px to tighten vertical space
      }}
    >
      {/* ── Left Content Column ── */}
      <div
        style={{
          padding: "64px 64px 48px 56px", // Tightened vertical padding
          borderRight: "1px solid var(--border)",
          display: "flex",
          flexDirection: "column",
          justifyContent: "flex-start", // Changed from space-between to eliminate center gaps
        }}
      >
        <div style={{ marginBottom: "auto" }}>
          {" "}
          {/* Pushes stats to bottom but keeps header tight */}
          <div
            className="fade-up"
            style={{
              display: "flex",
              alignItems: "center",
              gap: "12px",
              marginBottom: "24px", // Reduced from 32px
            }}
          >
            <span
              style={{
                width: "24px",
                height: "2px",
                background: "var(--accent)",
              }}
            />
            <span
              style={{
                ...sans,
                fontSize: "11px",
                color: "var(--accent)",
                fontWeight: 700,
                letterSpacing: "0.1em",
                textTransform: "uppercase",
              }}
            >
              Exchange Infrastructure Service
            </span>
          </div>
          <h1
            className="fade-up-1"
            style={{
              ...title,
              fontSize: "clamp(48px, 5.5vw, 82px)", // Slightly smaller to reduce vertical footprint
              lineHeight: "0.95",
              color: "var(--navy)",
              marginBottom: "32px", // Reduced from 40px
            }}
          >
            THE ENGINE OF <br />
            <span style={{ color: "var(--border)" }}>MARKET FINALITY.</span>
          </h1>
          <p
            className="fade-up-2"
            style={{
              ...sans,
              fontSize: "15px", // Tightened font size
              fontWeight: 400,
              color: "var(--text-muted)",
              lineHeight: "1.6",
              maxWidth: "480px",
              marginBottom: "40px", // Reduced from 48px
            }}
          >
            ESX owns the full trade lifecycle end to end. The engine that pairs
            buyers with sellers, the clearing house that guarantees every trade,
            and the ledger that records every movement with double-entry
            precision.
          </p>
          <div className="fade-up-3" style={{ display: "flex", gap: "12px" }}>
            <a
              href="#"
              style={{
                ...sans,
                fontSize: "12px",
                fontWeight: 600,
                textTransform: "uppercase",
                letterSpacing: "0.05em",
                background: "var(--accent)",
                color: "#ffffff",
                textDecoration: "none",
                padding: "14px 28px",
                display: "inline-flex",
                alignItems: "center",
                gap: "12px",
                transition: "background 0.15s",
              }}
              onMouseEnter={(e) =>
                (e.currentTarget.style.background = "var(--navy)")
              }
              onMouseLeave={(e) =>
                (e.currentTarget.style.background = "var(--accent)")
              }
            >
              Execute Market Entry <span>→</span>
            </a>
            <a
              href="#"
              style={{
                ...sans,
                fontSize: "12px",
                fontWeight: 600,
                textTransform: "uppercase",
                letterSpacing: "0.05em",
                border: "1px solid var(--border)",
                color: "var(--navy)",
                textDecoration: "none",
                padding: "14px 28px",
                transition: "border-color 0.15s",
              }}
              onMouseEnter={(e) =>
                (e.currentTarget.style.borderColor = "var(--navy)")
              }
              onMouseLeave={(e) =>
                (e.currentTarget.style.borderColor = "var(--border)")
              }
            >
              View Documentation
            </a>
          </div>
        </div>

        {/* Stats — Tightened Bottom Section */}
        <div
          className="fade-up-4"
          style={{
            display: "grid",
            gridTemplateColumns: "repeat(4, 1fr)",
            gap: "0",
            marginTop: "48px", // Reduced from 64px
            paddingTop: "32px", // Reduced from 40px
            borderTop: "1px solid var(--border)",
          }}
        >
          {[
            { v: "< 840µs", l: "Latency" }, // Shortened labels
            { v: "FIX 4.2", l: "Protocol" },
            { v: "Atomic", l: "Settlement" },
            { v: "Tier 1", l: "Liquidity" },
          ].map((s, i) => (
            <div
              key={s.l}
              style={{
                paddingRight: "16px",
                borderRight: i < 3 ? "1px solid var(--border)" : "none",
                paddingLeft: i > 0 ? "16px" : "0",
              }}
            >
              <div
                style={{
                  ...title,
                  fontSize: "18px",
                  color: "var(--navy)",
                  marginBottom: "4px",
                }}
              >
                {s.v}
              </div>
              <div
                style={{
                  ...sans,
                  fontSize: "9px",
                  color: "var(--text-subtle)",
                  fontWeight: 700,
                  textTransform: "uppercase",
                  letterSpacing: "0.05em",
                }}
              >
                {s.l}
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* ── Right Terminal: High Density Panel ── */}
      <div
        className="fade-up-2"
        style={{
          background: "var(--bg-deep)",
          display: "flex",
          flexDirection: "column",
          width: "100%",
        }}
      >
        {/* Panel Header */}
        <div
          style={{
            padding: "16px 24px",
            background: "var(--navy)",
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
            flexShrink: 0,
          }}
        >
          <div style={{ display: "flex", alignItems: "baseline", gap: "12px" }}>
            <span
              style={{
                ...sans,
                fontSize: "14px",
                fontWeight: 800,
                color: "#ffffff",
                letterSpacing: "0.02em",
                textTransform: "uppercase",
              }}
            >
              RELIANCE
            </span>
            <span
              style={{
                ...sans,
                fontSize: "10px",
                color: "rgba(255,255,255,0.4)",
                textTransform: "uppercase",
                letterSpacing: "0.1em",
                fontWeight: 600,
              }}
            >
              NSE · EQ
            </span>
          </div>
          <div style={{ display: "flex", alignItems: "center", gap: "12px" }}>
            <span
              style={{
                ...mono,
                fontSize: "18px",
                fontWeight: 600,
                color: "#ffffff",
              }}
            >
              2,847.50
            </span>
            <span
              style={{
                ...mono,
                fontSize: "10px",
                fontWeight: 700,
                color: "#ffffff",
                background: "var(--up)",
                padding: "2px 8px",
              }}
            >
              +1.24%
            </span>
          </div>
        </div>

        {/* Day Stats: Tighter Grid */}
        <div
          style={{
            display: "grid",
            gridTemplateColumns: "repeat(4, 1fr)",
            borderBottom: "1px solid var(--border)",
            flexShrink: 0,
            background: "#ffffff",
          }}
        >
          {[
            { l: "Open", v: "2,812.30" },
            { l: "High", v: "2,861.00" },
            { l: "Low", v: "2,808.75" },
            { l: "Volume", v: "4.2M" },
          ].map((d, i) => (
            <div
              key={d.l}
              style={{
                padding: "12px 24px",
                borderRight: i < 3 ? "1px solid var(--border)" : "none",
              }}
            >
              <div
                style={{
                  ...sans,
                  fontSize: "9px",
                  color: "var(--text-subtle)",
                  textTransform: "uppercase",
                  fontWeight: 700,
                  marginBottom: "2px",
                }}
              >
                {d.l}
              </div>
              <div
                style={{
                  ...mono,
                  fontSize: "12px",
                  color: "var(--navy)",
                  fontWeight: 600,
                }}
              >
                {d.v}
              </div>
            </div>
          ))}
        </div>

        {/* Order Book: Condensed rows */}
        <div
          style={{
            flex: 1,
            display: "flex",
            flexDirection: "column",
            overflow: "hidden",
            background: "#ffffff",
          }}
        >
          <div
            style={{
              display: "grid",
              gridTemplateColumns: "1fr 1fr 1fr",
              padding: "8px 24px",
              borderBottom: "1px solid var(--border)",
              background: "var(--bg-deep)",
            }}
          >
            {["Price", "Qty", "Value"].map((h) => (
              <span
                key={h}
                style={{
                  ...sans,
                  fontSize: "9px",
                  fontWeight: 700,
                  textTransform: "uppercase",
                  color: "var(--text-muted)",
                  textAlign:
                    h === "Qty" ? "center" : h === "Value" ? "right" : "left",
                }}
              >
                {h}
              </span>
            ))}
          </div>

          {/* Asks */}
          <div style={{ padding: "4px 0" }}>
            {asks.slice(0, 4).map(
              (
                r, // Showing 4 levels for height efficiency
              ) => (
                <div
                  key={r.p}
                  style={{
                    display: "grid",
                    gridTemplateColumns: "1fr 1fr 1fr",
                    padding: "4px 24px",
                    position: "relative",
                  }}
                >
                  <div
                    style={{
                      position: "absolute",
                      right: 0,
                      top: 0,
                      bottom: 0,
                      width: `${r.d}%`,
                      background: "rgba(220, 38, 38, 0.05)",
                    }}
                  />
                  <span
                    style={{
                      ...mono,
                      fontSize: "12px",
                      color: "var(--down)",
                      fontWeight: 600,
                      position: "relative",
                    }}
                  >
                    {r.p}
                  </span>
                  <span
                    style={{
                      ...mono,
                      fontSize: "12px",
                      color: "var(--navy)",
                      textAlign: "center",
                      position: "relative",
                    }}
                  >
                    {r.q}
                  </span>
                  <span
                    style={{
                      ...mono,
                      fontSize: "11px",
                      color: "var(--text-muted)",
                      textAlign: "right",
                      position: "relative",
                    }}
                  >
                    {r.v}
                  </span>
                </div>
              ),
            )}
          </div>

          {/* Tighter Spread Bar */}
          <div
            style={{
              display: "flex",
              gap: "10px",
              padding: "6px 24px",
              borderTop: "1px solid var(--border)",
              borderBottom: "1px solid var(--border)",
              background: "var(--bg-deep)",
            }}
          >
            <span
              style={{
                ...sans,
                fontSize: "9px",
                fontWeight: 700,
                textTransform: "uppercase",
                color: "var(--text-muted)",
              }}
            >
              Spread
            </span>
            <span
              style={{
                ...mono,
                fontSize: "12px",
                color: "var(--navy)",
                fontWeight: 700,
              }}
            >
              0.25
            </span>
            <span
              style={{
                ...sans,
                fontSize: "9px",
                fontWeight: 700,
                color: "var(--text-subtle)",
                marginLeft: "auto",
              }}
            >
              L2 Visibility
            </span>
          </div>

          {/* Bids */}
          <div style={{ padding: "4px 0" }}>
            {bids.slice(0, 4).map((r) => (
              <div
                key={r.p}
                style={{
                  display: "grid",
                  gridTemplateColumns: "1fr 1fr 1fr",
                  padding: "4px 24px",
                  position: "relative",
                }}
              >
                <div
                  style={{
                    position: "absolute",
                    right: 0,
                    top: 0,
                    bottom: 0,
                    width: `${r.d}%`,
                    background: "rgba(22, 163, 74, 0.05)",
                  }}
                />
                <span
                  style={{
                    ...mono,
                    fontSize: "12px",
                    color: "var(--up)",
                    fontWeight: 600,
                    position: "relative",
                  }}
                >
                  {r.p}
                </span>
                <span
                  style={{
                    ...mono,
                    fontSize: "12px",
                    color: "var(--navy)",
                    textAlign: "center",
                    position: "relative",
                  }}
                >
                  {r.q}
                </span>
                <span
                  style={{
                    ...mono,
                    fontSize: "11px",
                    color: "var(--text-muted)",
                    textAlign: "right",
                    position: "relative",
                  }}
                >
                  {r.v}
                </span>
              </div>
            ))}
          </div>
        </div>

        {/* Recent Trades: Compact view */}
        <div
          style={{
            borderTop: "1px solid var(--border)",
            background: "#ffffff",
          }}
        >
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              padding: "8px 24px",
              background: "var(--bg-deep)",
              borderBottom: "1px solid var(--border)",
            }}
          >
            <span
              style={{
                ...sans,
                fontSize: "9px",
                fontWeight: 700,
                textTransform: "uppercase",
                color: "var(--text-muted)",
              }}
            >
              Executions
            </span>
            <span
              style={{
                ...mono,
                fontSize: "10px",
                color: "var(--text-subtle)",
                fontWeight: 700,
              }}
            >
              {time}
            </span>
          </div>
          {trades.slice(0, 3).map((t, i) => (
            <div
              key={i}
              style={{
                display: "grid",
                gridTemplateColumns: "1fr 1fr 1fr 60px",
                padding: "6px 24px",
                borderBottom: i < 2 ? "1px solid var(--border-light)" : "none",
              }}
            >
              <span
                style={{
                  ...mono,
                  fontSize: "11px",
                  color: "var(--text-muted)",
                }}
              >
                {t.t}
              </span>
              <span
                style={{
                  ...mono,
                  fontSize: "11px",
                  fontWeight: 600,
                  color: t.side ? "var(--up)" : "var(--down)",
                }}
              >
                {t.p}
              </span>
              <span
                style={{
                  ...mono,
                  fontSize: "11px",
                  color: "var(--navy)",
                  textAlign: "center",
                }}
              >
                {t.q}
              </span>
              <span
                style={{
                  ...sans,
                  fontSize: "9px",
                  fontWeight: 800,
                  textTransform: "uppercase",
                  color: t.side ? "var(--up)" : "var(--down)",
                  textAlign: "right",
                }}
              >
                {t.side ? "Buy" : "Sell"}
              </span>
            </div>
          ))}
        </div>

        {/* Footer Bar */}
        <div
          style={{
            padding: "10px 24px",
            background: "var(--navy)",
            display: "flex",
            justifyContent: "space-between",
          }}
        >
          <span
            style={{
              ...sans,
              fontSize: "9px",
              fontWeight: 700,
              textTransform: "uppercase",
              color: "rgba(255,255,255,0.4)",
            }}
          >
            Price-Time Priority
          </span>
          <span
            style={{
              ...sans,
              fontSize: "9px",
              fontWeight: 800,
              textTransform: "uppercase",
              color: "var(--accent)",
            }}
          >
            DvP Finality
          </span>
        </div>
      </div>
    </section>
  );
}

/* ─── Services ──────────────────────────────────────────────── */

function Services() {
  return (
    <section
      style={{ background: "#ffffff", borderBottom: "1px solid var(--border)" }}
    >
      {/* Header */}
      <div
        style={{
          padding: "72px 56px 48px",
          borderBottom: "1px solid var(--border)",
          display: "flex",
          alignItems: "flex-end",
          justifyContent: "space-between",
          gap: "40px",
        }}
      >
        <div>
          <p
            style={{
              ...sans,
              fontSize: "11px",
              color: "var(--accent)",
              fontWeight: 700,
              textTransform: "uppercase",
              letterSpacing: "0.1em",
              marginBottom: "16px",
            }}
          >
            Exchange Infrastructure
          </p>
          <h2
            style={{
              ...title,
              fontSize: "clamp(32px, 3vw, 48px)",
              color: "var(--navy)",
              lineHeight: 1.1,
            }}
          >
            FROM EXECUTION <br /> TO FINALITY.
          </h2>
        </div>
        <p
          style={{
            ...sans,
            fontSize: "15px",
            fontWeight: 400,
            color: "var(--text-muted)",
            maxWidth: "420px",
            lineHeight: "1.7",
          }}
        >
          ESX handles every stage from the moment an order is submitted to the
          moment the ledger entry is written — no external dependencies, no
          partial coverage.
        </p>
      </div>

      {/* Grid */}
      <div style={{ display: "grid", gridTemplateColumns: "repeat(3, 1fr)" }}>
        {SERVICES.map((s, i) => (
          <div
            key={s.n}
            style={{
              padding: "48px",
              borderRight:
                (i + 1) % 3 !== 0 ? "1px solid var(--border)" : "none",
              borderBottom: i < 3 ? "1px solid var(--border)" : "none",
              transition: "background 0.2s",
            }}
            onMouseEnter={(e) =>
              (e.currentTarget.style.background = "var(--bg-deep)")
            }
            onMouseLeave={(e) =>
              (e.currentTarget.style.background = "transparent")
            }
          >
            <div
              style={{
                display: "flex",
                justifyContent: "space-between",
                alignItems: "flex-start",
                marginBottom: "32px",
              }}
            >
              <span
                style={{ ...title, fontSize: "14px", color: "var(--accent)" }}
              >
                {s.n}
              </span>
              <div style={{ textAlign: "right" }}>
                <div
                  style={{
                    ...title,
                    fontSize: "24px",
                    color: "var(--navy)",
                    lineHeight: 1,
                  }}
                >
                  {s.metric}
                </div>
                <div
                  style={{
                    ...sans,
                    fontSize: "9px",
                    color: "var(--text-muted)",
                    marginTop: "6px",
                    fontWeight: 600,
                    textTransform: "uppercase",
                    letterSpacing: "0.05em",
                  }}
                >
                  {s.mLabel}
                </div>
              </div>
            </div>
            <h3
              style={{
                ...title,
                fontSize: "16px",
                color: "var(--navy)",
                marginBottom: "16px",
              }}
            >
              {s.title}
            </h3>
            <p
              style={{
                ...sans,
                fontSize: "14px",
                fontWeight: 400,
                color: "var(--text-muted)",
                lineHeight: "1.7",
              }}
            >
              {s.body}
            </p>
          </div>
        ))}
      </div>
    </section>
  );
}

/* ─── Lifecycle ─────────────────────────────────────────────── */

function Lifecycle() {
  return (
    <section
      style={{
        background: "var(--navy)",
        borderBottom: "1px solid rgba(255,255,255,0.1)",
        padding: "88px 56px",
      }}
    >
      <div
        style={{
          display: "flex",
          alignItems: "flex-end",
          justifyContent: "space-between",
          marginBottom: "72px",
        }}
      >
        <div>
          <p
            style={{
              ...sans,
              fontSize: "11px",
              color: "var(--accent)",
              fontWeight: 700,
              textTransform: "uppercase",
              letterSpacing: "0.1em",
              marginBottom: "16px",
            }}
          >
            The Trade Lifecycle
          </p>
          <h2
            style={{
              ...title,
              fontSize: "clamp(32px, 3vw, 48px)",
              color: "#ffffff",
              lineHeight: 1.1,
            }}
          >
            SIX STEPS TO <br /> ABSOLUTE FINALITY.
          </h2>
        </div>
        <a
          href="#"
          style={{
            ...sans,
            fontSize: "12px",
            fontWeight: 600,
            textTransform: "uppercase",
            letterSpacing: "0.05em",
            color: "#ffffff",
            textDecoration: "none",
            border: "1px solid rgba(255,255,255,0.2)",
            padding: "14px 28px",
            borderRadius: "0px",
            transition: "all 0.15s",
          }}
          onMouseEnter={(e) => {
            e.currentTarget.style.background = "#ffffff";
            e.currentTarget.style.color = "var(--navy)";
          }}
          onMouseLeave={(e) => {
            e.currentTarget.style.background = "transparent";
            e.currentTarget.style.color = "#ffffff";
          }}
        >
          View Documentation
        </a>
      </div>

      <div
        style={{
          display: "grid",
          gridTemplateColumns: "repeat(6, 1fr)",
          gap: "0",
          position: "relative",
        }}
      >
        {/* Connector */}
        <div
          style={{
            position: "absolute",
            top: "17px",
            left: "3%",
            right: "3%",
            height: "1px",
            background: "rgba(255,255,255,0.1)",
          }}
        />

        {LIFECYCLE.map((s, i) => (
          <div
            key={s.step}
            style={{
              paddingRight: i < 5 ? "32px" : "0",
              position: "relative",
              zIndex: 1,
            }}
          >
            <div
              style={{
                width: "36px",
                height: "36px",
                border: "1px solid rgba(255,255,255,0.2)",
                background: "var(--navy)",
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
                marginBottom: "24px",
                borderRadius: "0px",
              }}
            >
              <span
                style={{
                  ...title,
                  fontSize: "12px",
                  color: "var(--accent)",
                }}
              >
                {s.step}
              </span>
            </div>
            <h4
              style={{
                ...title,
                fontSize: "14px",
                color: "#ffffff",
                marginBottom: "12px",
              }}
            >
              {s.title}
            </h4>
            <p
              style={{
                ...sans,
                fontSize: "13px",
                fontWeight: 400,
                color: "rgba(255,255,255,0.6)",
                lineHeight: "1.7",
              }}
            >
              {s.desc}
            </p>
          </div>
        ))}
      </div>
    </section>
  );
}

/* ─── Membership ────────────────────────────────────────────── */

function Membership() {
  return (
    <section
      style={{ background: "#ffffff", borderBottom: "1px solid var(--border)" }}
    >
      <div
        style={{
          padding: "72px 56px 0",
          borderBottom: "1px solid var(--border)",
          paddingBottom: "40px",
          display: "flex",
          alignItems: "flex-end",
          justifyContent: "space-between",
        }}
      >
        <div>
          <p
            style={{
              ...sans,
              fontSize: "11px",
              color: "var(--accent)",
              fontWeight: 700,
              textTransform: "uppercase",
              letterSpacing: "0.1em",
              marginBottom: "16px",
            }}
          >
            Participation Access
          </p>
          <h2
            style={{
              ...title,
              fontSize: "clamp(32px, 3vw, 48px)",
              color: "var(--navy)",
            }}
          >
            BUILT FOR PROFESSIONALS.
          </h2>
        </div>
        <a
          href="#"
          style={{
            ...sans,
            fontSize: "12px",
            fontWeight: 600,
            textTransform: "uppercase",
            letterSpacing: "0.05em",
            background: "var(--navy)",
            color: "#ffffff",
            textDecoration: "none",
            padding: "14px 28px",
            borderRadius: "0px",
            transition: "background 0.15s",
          }}
          onMouseEnter={(e) =>
            (e.currentTarget.style.background = "var(--accent)")
          }
          onMouseLeave={(e) =>
            (e.currentTarget.style.background = "var(--navy)")
          }
        >
          View Member Criteria
        </a>
      </div>

      <div style={{ display: "grid", gridTemplateColumns: "repeat(4, 1fr)" }}>
        {MEMBERS.map((m, i) => (
          <div
            key={m.title}
            style={{
              padding: "48px 40px",
              borderRight: i < 3 ? "1px solid var(--border)" : "none",
            }}
          >
            <div style={{ marginBottom: "24px" }}>
              <div
                style={{
                  ...title,
                  fontSize: "16px",
                  color: "var(--navy)",
                  marginBottom: "6px",
                }}
              >
                {m.title}
              </div>
              <div
                style={{
                  ...sans,
                  fontSize: "11px",
                  color: "var(--text-muted)",
                  fontWeight: 600,
                  textTransform: "uppercase",
                  letterSpacing: "0.05em",
                }}
              >
                {m.sub}
              </div>
            </div>
            <ul
              style={{
                listStyle: "none",
                display: "flex",
                flexDirection: "column",
                gap: "12px",
              }}
            >
              {m.items.map((item) => (
                <li
                  key={item}
                  style={{
                    display: "flex",
                    gap: "12px",
                    alignItems: "flex-start",
                  }}
                >
                  <span
                    style={{
                      color: "var(--accent)",
                      marginTop: "2px",
                      fontSize: "10px",
                      flexShrink: 0,
                    }}
                  >
                    ■
                  </span>
                  <span
                    style={{
                      ...sans,
                      fontSize: "13px",
                      fontWeight: 400,
                      color: "var(--text-muted)",
                      lineHeight: "1.5",
                    }}
                  >
                    {item}
                  </span>
                </li>
              ))}
            </ul>
          </div>
        ))}
      </div>
    </section>
  );
}

/* ─── Announcement bar / News strip ────────────────────────── */

function NewsStrip() {
  const news = [
    {
      tag: "Notice",
      text: "Settlement calendar updated for upcoming public holidays. See the exchange notice for revised schedules.",
    },
    {
      tag: "Listing",
      text: "New securities available for trading from Monday, 5 May 2026. Review the listing circular.",
    },
    {
      tag: "Circular",
      text: "Revised margin requirements for F&O segment effective 1 June 2026. Download the risk framework update.",
    },
  ];

  return (
    <section
      style={{
        background: "var(--bg-deep)",
        borderBottom: "1px solid var(--border)",
      }}
    >
      <div
        style={{
          padding: "0 56px",
          borderBottom: "1px solid var(--border)",
          height: "48px",
          display: "flex",
          alignItems: "center",
        }}
      >
        <span
          style={{
            ...sans,
            fontSize: "11px",
            color: "var(--navy)",
            fontWeight: 700,
            textTransform: "uppercase",
            letterSpacing: "0.1em",
          }}
        >
          Regulatory Notices
        </span>
      </div>
      <div>
        {news.map((n, i) => (
          <div
            key={i}
            style={{
              padding: "20px 56px",
              borderBottom:
                i < news.length - 1 ? "1px solid var(--border)" : "none",
              display: "flex",
              alignItems: "center",
              gap: "24px",
              transition: "background 0.15s",
              cursor: "pointer",
            }}
            onMouseEnter={(e) => (e.currentTarget.style.background = "#ffffff")}
            onMouseLeave={(e) =>
              (e.currentTarget.style.background = "transparent")
            }
          >
            <span
              style={{
                ...title,
                fontSize: "10px",
                color: "var(--navy)",
                border: "1px solid var(--border)",
                background: "#ffffff",
                padding: "6px 12px",
                whiteSpace: "nowrap",
              }}
            >
              {n.tag}
            </span>
            <span
              style={{
                ...sans,
                fontSize: "14px",
                fontWeight: 400,
                color: "var(--text-muted)",
              }}
            >
              {n.text}
            </span>
            <span
              style={{
                ...sans,
                fontSize: "12px",
                fontWeight: 600,
                textTransform: "uppercase",
                letterSpacing: "0.05em",
                color: "var(--accent)",
                marginLeft: "auto",
                whiteSpace: "nowrap",
              }}
            >
              Read Notice →
            </span>
          </div>
        ))}
      </div>
    </section>
  );
}

/* ─── Compliance ────────────────────────────────────────────── */

function Compliance() {
  const items = [
    {
      title: "Central Counterparty",
      desc: "ESX interposes as CCP on every executed trade. Neither buyer nor seller carries exposure to the other — only to ESX, which guarantees final settlement.",
    },
    {
      title: "Circuit Breakers",
      desc: "Trading in any security is automatically halted when the price moves more than 10% within a 60-second window. Identical to the mechanisms used by NSE, BSE, and NASDAQ.",
    },
    {
      title: "Pre-Trade Validation",
      desc: "No order enters the matching engine without a verified and locked collateral position. The risk engine rejects deficient orders before they are ever placed in the book.",
    },
    {
      title: "Immutable Audit Trail",
      desc: "Every order, match, clearing event, and settlement leg is timestamped at microsecond precision and stored permanently. Full reconstruction of any transaction is possible at any time.",
    },
    {
      title: "Double-Entry Ledger",
      desc: "All movements are recorded as balanced journal entries. The ledger cannot be modified unilaterally. Debits equal credits at every point in time, across both cash and securities dimensions.",
    },
    {
      title: "Regulatory Reporting",
      desc: "ESX produces the structured reports required by market regulators — trade repository submissions, position reports, settlement confirmations, and reconciliation statements.",
    },
  ];

  return (
    <section
      style={{ background: "#ffffff", borderBottom: "1px solid var(--border)" }}
    >
      <div
        style={{
          padding: "72px 56px 40px",
          borderBottom: "1px solid var(--border)",
          display: "flex",
          alignItems: "flex-end",
          justifyContent: "space-between",
        }}
      >
        <div>
          <p
            style={{
              ...sans,
              fontSize: "11px",
              color: "var(--accent)",
              fontWeight: 700,
              textTransform: "uppercase",
              letterSpacing: "0.1em",
              marginBottom: "16px",
            }}
          >
            Risk & Compliance
          </p>
          <h2
            style={{
              ...title,
              fontSize: "clamp(32px, 3vw, 48px)",
              color: "var(--navy)",
            }}
          >
            EXCHANGE-GRADE CONTROLS.
          </h2>
        </div>
        <a
          href="#"
          style={{
            ...sans,
            fontSize: "12px",
            fontWeight: 600,
            textTransform: "uppercase",
            letterSpacing: "0.05em",
            color: "var(--navy)",
            textDecoration: "none",
            borderBottom: "2px solid var(--accent)",
            paddingBottom: "4px",
          }}
        >
          Download Rulebook
        </a>
      </div>

      <div style={{ display: "grid", gridTemplateColumns: "repeat(3, 1fr)" }}>
        {items.map((item, i) => (
          <div
            key={item.title}
            style={{
              padding: "40px 48px",
              borderRight:
                (i + 1) % 3 !== 0 ? "1px solid var(--border)" : "none",
              borderBottom: i < 3 ? "1px solid var(--border)" : "none",
            }}
          >
            <h4
              style={{
                ...title,
                fontSize: "15px",
                color: "var(--navy)",
                marginBottom: "12px",
              }}
            >
              {item.title}
            </h4>
            <p
              style={{
                ...sans,
                fontSize: "14px",
                fontWeight: 400,
                color: "var(--text-muted)",
                lineHeight: "1.7",
              }}
            >
              {item.desc}
            </p>
          </div>
        ))}
      </div>
    </section>
  );
}

/* ─── CTA ───────────────────────────────────────────────────── */

function CTA() {
  return (
    <section
      style={{
        background: "var(--navy)",
        padding: "100px 56px",
        borderBottom: "1px solid rgba(255,255,255,0.1)",
        display: "flex",
        alignItems: "center",
        justifyContent: "space-between",
        gap: "60px",
      }}
    >
      <div style={{ maxWidth: "600px" }}>
        <p
          style={{
            ...sans,
            fontSize: "11px",
            color: "var(--accent)",
            fontWeight: 700,
            textTransform: "uppercase",
            letterSpacing: "0.1em",
            marginBottom: "20px",
          }}
        >
          Exchange Connectivity
        </p>
        <h2
          style={{
            ...title,
            fontSize: "clamp(40px, 4vw, 64px)",
            color: "#ffffff",
            lineHeight: "1",
            marginBottom: "24px",
          }}
        >
          READY FOR <br />
          EXECUTION.
        </h2>
        <p
          style={{
            ...sans,
            fontSize: "16px",
            fontWeight: 400,
            color: "rgba(255,255,255,0.7)",
            lineHeight: "1.7",
          }}
        >
          Brokers, institutions, algorithmic trading firms, and custodians can
          apply for ESX membership and connect directly via FIX 4.2.
        </p>
      </div>

      <div
        style={{
          display: "flex",
          flexDirection: "column",
          gap: "16px",
          flexShrink: 0,
        }}
      >
        <a
          href="#"
          style={{
            ...sans,
            fontSize: "13px",
            fontWeight: 700,
            textTransform: "uppercase",
            letterSpacing: "0.05em",
            background: "var(--accent)",
            color: "#ffffff",
            textDecoration: "none",
            padding: "18px 48px",
            textAlign: "center",
            whiteSpace: "nowrap",
            borderRadius: "0px",
            transition: "background 0.15s",
          }}
          onMouseEnter={(e) => (
            (e.currentTarget.style.background = "#ffffff"),
            (e.currentTarget.style.color = "var(--navy)")
          )}
          onMouseLeave={(e) => (
            (e.currentTarget.style.background = "var(--accent)"),
            (e.currentTarget.style.color = "#ffffff")
          )}
        >
          Request FIX Credentials
        </a>
        <a
          href="#"
          style={{
            ...sans,
            fontSize: "13px",
            fontWeight: 600,
            textTransform: "uppercase",
            letterSpacing: "0.05em",
            border: "1px solid rgba(255,255,255,0.2)",
            color: "#ffffff",
            textDecoration: "none",
            padding: "18px 48px",
            textAlign: "center",
            whiteSpace: "nowrap",
            borderRadius: "0px",
            transition: "background 0.15s",
          }}
          onMouseEnter={(e) =>
            (e.currentTarget.style.background = "rgba(255,255,255,0.05)")
          }
          onMouseLeave={(e) =>
            (e.currentTarget.style.background = "transparent")
          }
        >
          Read Specifications
        </a>
      </div>
    </section>
  );
}

/* ─── Footer ────────────────────────────────────────────────── */

function Footer() {
  const cols = [
    {
      h: "Markets",
      links: ["Equities", "Derivatives", "Fixed Income", "Currency", "ETFs"],
    },
    {
      h: "Participants",
      links: [
        "Trading Members",
        "Clearing Members",
        "Custodians",
        "Issuers",
        "Data Vendors",
      ],
    },
    {
      h: "Exchange",
      links: [
        "Rules & Regulations",
        "Listing Requirements",
        "Fee Schedule",
        "Settlement Calendar",
        "Circuit Breaker Policy",
      ],
    },
    {
      h: "Technology",
      links: [
        "FIX Connectivity",
        "Market Data API",
        "WebSocket Feeds",
        "Co-location",
        "Documentation",
      ],
    },
    {
      h: "About",
      links: ["About ESX", "Leadership", "Press & Media", "Careers", "Contact"],
    },
  ];

  return (
    <footer
      style={{
        background: "var(--bg-deep)",
        borderTop: "1px solid var(--border)",
      }}
    >
      <div
        style={{
          padding: "72px 56px",
          display: "grid",
          gridTemplateColumns: "1.4fr repeat(5, 1fr)",
          gap: "0",
          borderBottom: "1px solid var(--border)",
        }}
      >
        {/* Brand */}
        <div style={{ paddingRight: "40px" }}>
          <div
            style={{
              display: "flex",
              alignItems: "center",
              gap: "12px",
              marginBottom: "16px",
            }}
          >
            <div
              style={{
                width: "32px",
                height: "32px",
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
              }}
            >
              <img
                src="/logo.png"
                alt="ESX Logo"
                style={{
                  width: "100%",
                  height: "100%",
                  objectFit: "contain",
                }}
              />
            </div>
            <span
              style={{
                ...title,
                fontSize: "16px",
                color: "var(--navy)",
              }}
            >
              ESX
            </span>
          </div>
          <p
            style={{
              ...sans,
              fontSize: "13px",
              fontWeight: 400,
              color: "var(--text-muted)",
              lineHeight: "1.7",
              maxWidth: "240px",
            }}
          >
            Escrow Stock Exchange. Production-grade securities exchange
            infrastructure built for high-frequency liquidity.
          </p>
        </div>

        {cols.map((col) => (
          <div key={col.h}>
            <div
              style={{
                ...sans,
                fontSize: "11px",
                fontWeight: 700,
                color: "var(--navy)",
                marginBottom: "20px",
                textTransform: "uppercase",
                letterSpacing: "0.05em",
              }}
            >
              {col.h}
            </div>
            <ul
              style={{
                listStyle: "none",
                display: "flex",
                flexDirection: "column",
                gap: "12px",
              }}
            >
              {col.links.map((link) => (
                <li key={link}>
                  <a
                    href="#"
                    style={{
                      ...sans,
                      fontSize: "13px",
                      fontWeight: 400,
                      color: "var(--text-muted)",
                      textDecoration: "none",
                      transition: "color 0.15s",
                    }}
                    onMouseEnter={(e) =>
                      (e.currentTarget.style.color = "var(--accent)")
                    }
                    onMouseLeave={(e) =>
                      (e.currentTarget.style.color = "var(--text-muted)")
                    }
                  >
                    {link}
                  </a>
                </li>
              ))}
            </ul>
          </div>
        ))}
      </div>

      {/* Bottom */}
      <div
        style={{
          padding: "24px 56px",
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
        }}
      >
        <span
          style={{
            ...sans,
            fontSize: "11px",
            color: "var(--text-subtle)",
            fontWeight: 600,
            textTransform: "uppercase",
            letterSpacing: "0.05em",
          }}
        >
          © 2026 Escrow Stock Exchange.
        </span>
        <div style={{ display: "flex", gap: "32px" }}>
          {[
            "Privacy Policy",
            "Terms of Use",
            "Disclosures",
            "Connectivity Specs",
          ].map((l) => (
            <a
              key={l}
              href="#"
              style={{
                ...sans,
                fontSize: "11px",
                color: "var(--text-subtle)",
                textDecoration: "none",
                fontWeight: 600,
                textTransform: "uppercase",
                letterSpacing: "0.05em",
                transition: "color 0.15s",
              }}
              onMouseEnter={(e) =>
                (e.currentTarget.style.color = "var(--navy)")
              }
              onMouseLeave={(e) =>
                (e.currentTarget.style.color = "var(--text-subtle)")
              }
            >
              {l}
            </a>
          ))}
        </div>
      </div>
    </footer>
  );
}

/* ─── Page ──────────────────────────────────────────────────── */

export default function Page() {
  return (
    <main style={{ width: "100%", overflowX: "hidden" }}>
      <Ticker />
      <Nav />
      <IndexBar />
      <Hero />
      <Services />
      <Lifecycle />
      <Membership />
      <NewsStrip />
      <Compliance />
      <CTA />
      <Footer />
    </main>
  );
}
