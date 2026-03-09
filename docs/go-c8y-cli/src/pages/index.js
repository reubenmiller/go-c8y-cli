import React, { useState, useEffect } from 'react';
import clsx from 'clsx';
import Layout from '@theme/Layout';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import useBaseUrl from '@docusaurus/useBaseUrl';
import CookieConsent from "react-cookie-consent";
import AsciinemaPlayer from '@site/src/components/AsciinemaPlayer';

import styles from './styles.module.css';

// Animation: "Cumulocity" → select all → Transform to Lowercase → cumulocity
//           → backspace middle → c8y → reverse → select → Transform to Title Case → loop
const FRAMES = (() => {
  const TYPING = 75;
  const PAUSE  = 2200;
  const SELECT = 380; // brief pause while text appears selected
  const mid    = 'umulocit';
  const f      = [];

  // Show full title normally
  f.push({ text: 'Cumulocity', delay: PAUSE });

  // Select all (Ctrl+A) — text appears highlighted
  f.push({ text: 'Cumulocity', delay: SELECT, selected: true });

  // "Transform to Lowercase" command applied while still selected
  f.push({ text: 'cumulocity', delay: SELECT, selected: true });

  // Deselect — cursor placed, ready to edit
  f.push({ text: 'cumulocity', delay: TYPING });

  // Backspace 'umulocit' one char at a time from the right (before 'y')
  for (let i = mid.length - 1; i >= 0; i--)
    f.push({ text: `c${mid.slice(0, i)}y`, delay: TYPING });

  // Type '8' between 'c' and 'y'
  f.push({ text: 'c8y', delay: PAUSE + 1000 });

  // Reverse: delete '8' → 'cy'
  f.push({ text: 'cy', delay: TYPING });

  // Retype 'umulocit'
  for (let i = 1; i <= mid.length; i++)
    f.push({ text: `c${mid.slice(0, i)}y`, delay: TYPING });

  // Select 'cumulocity' again
  f.push({ text: 'cumulocity', delay: SELECT, selected: true });

  // "Transform to Title Case" — restores capital C while still selected
  f.push({ text: 'Cumulocity', delay: SELECT, selected: true });

  // Deselect — then loop back to frame 0 which shows 'Cumulocity' for PAUSE

  return f;
})();

function AnimatedTitle() {
  const [idx, setIdx] = useState(0);

  useEffect(() => {
    const t = setTimeout(() => setIdx(i => (i + 1) % FRAMES.length), FRAMES[idx].delay);
    return () => clearTimeout(t);
  }, [idx]);

  const { text, selected } = FRAMES[idx];
  const selectionStyle = selected
    ? { background: 'rgba(100, 160, 230, 0.4)', borderRadius: '3px', padding: '0 4px', margin: '0 -4px' }
    : {};

  return <span style={selectionStyle}>{text}</span>;
}

// Inline SVG icons — terminal/CLI themed
const IconAutomate = () => (
  <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" className={styles.featureIcon}>
    <rect x="2" y="3" width="20" height="15" rx="2" stroke="currentColor" strokeWidth="1.8" fill="none"/>
    <path d="M7 8l3 3-3 3" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"/>
    <path d="M13 14h4" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round"/>
    <path d="M6 21h12" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round"/>
  </svg>
);

const IconPipeline = () => (
  <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" className={styles.featureIcon}>
    <circle cx="4" cy="12" r="2.5" stroke="currentColor" strokeWidth="1.8"/>
    <circle cx="12" cy="12" r="2.5" stroke="currentColor" strokeWidth="1.8"/>
    <circle cx="20" cy="12" r="2.5" stroke="currentColor" strokeWidth="1.8"/>
    <path d="M6.5 12h3M14.5 12h3" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round"/>
    <path d="M4 6v3M12 6v3M20 6v3" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round"/>
    <path d="M4 15v3M12 15v3M20 15v3" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round"/>
  </svg>
);

const IconActivityLog = () => (
  <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" className={styles.featureIcon}>
    <rect x="3" y="3" width="18" height="18" rx="2" stroke="currentColor" strokeWidth="1.8" fill="none"/>
    <path d="M7 8h10" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round"/>
    <path d="M7 12h10" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round"/>
    <path d="M7 16h6" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round"/>
    <circle cx="19" cy="19" r="4" fill="var(--ifm-color-primary)" stroke="none"/>
    <path d="M17.5 19l1 1 2-2" stroke="white" strokeWidth="1.4" strokeLinecap="round" strokeLinejoin="round"/>
  </svg>
);

const IconWorkers = () => (
  <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" className={styles.featureIcon}>
    <path d="M3 6h3v12H3zM10.5 3h3v18h-3zM18 9h3v9h-3z" stroke="currentColor" strokeWidth="1.8" strokeLinejoin="round" fill="none"/>
  </svg>
);

const IconConfig = () => (
  <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" className={styles.featureIcon}>
    <path d="M4 6h2" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round"/>
    <circle cx="9" cy="6" r="2" stroke="currentColor" strokeWidth="1.8"/>
    <path d="M11 6h9" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round"/>
    <path d="M4 12h9" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round"/>
    <circle cx="15" cy="12" r="2" stroke="currentColor" strokeWidth="1.8"/>
    <path d="M17 12h3" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round"/>
    <path d="M4 18h2" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round"/>
    <circle cx="9" cy="18" r="2" stroke="currentColor" strokeWidth="1.8"/>
    <path d="M11 18h9" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round"/>
  </svg>
);

const features = [
  {
    title: 'Automate',
    Icon: IconAutomate,
    description: 'Write concise shell scripts to automate repetitive Cumulocity tasks — from device management to alarm handling.',
  },
  {
    title: 'Native pipeline support',
    Icon: IconPipeline,
    description: 'Chain commands together and pipe data into third-party tools. Compose complex workflows from simple building blocks.',
  },
  {
    title: 'Activity log',
    Icon: IconActivityLog,
    description: 'Every API interaction is recorded. Audit and replay requests to Cumulocity for full traceability.',
  },
  {
    title: 'Concurrent workers',
    Icon: IconWorkers,
    description: 'Control concurrency to process large lists of devices or operations in parallel without overwhelming the platform.',
  },
  {
    title: 'Highly configurable',
    Icon: IconConfig,
    description: 'Manage multiple sessions, customise output views, and tailor every aspect of the CLI to your workflow.',
  },
];

function Feature({ Icon, title, description }) {
  return (
    <div className={styles.featureCard}>
      <div className={styles.featureIconWrapper}>
        <Icon />
      </div>
      <h3 className={styles.featureTitle}>{title}</h3>
      <p className={styles.featureDesc}>{description}</p>
    </div>
  );
}

export default function Home() {
  const context = useDocusaurusContext();
  const {siteConfig = {}} = context;

  return (
    <Layout
      title={siteConfig.title}
      description="go-c8y-cli — Cumulocity IoT Command Line Interface">

      {/* ── Hero ─────────────────────────────────────────── */}
      <header className={styles.heroSection}>
        <a
          className={styles.heroBadge}
          href="https://github.com/reubenmiller/go-c8y-cli"
          target="_blank"
          rel="noopener noreferrer"
        >
          <span className={styles.heroBadgeDot} />
          Open source · MIT
        </a>
        <h1 className={styles.heroTitle}><AnimatedTitle /></h1>
        <p className={styles.heroSubtitle}>{siteConfig.tagline}</p>
        <div className={styles.heroButtons}>
          <Link className={styles.btnPrimary} to={useBaseUrl('docs/')}>
            Get started
          </Link>
          <Link className={styles.btnSecondary} to={useBaseUrl('docs/cli/')}>
            API reference
          </Link>
        </div>
        <div className={styles.demoWrapper}>
          <AsciinemaPlayer
            src="https://asciinema.org/a/413796.cast"
            cols={126}
            rows={30}
            preload
            fit="width"
            theme="monokai"
            poster="npt:0:03"
          />
        </div>
      </header>

      {/* ── Features ─────────────────────────────────────── */}
      <section className={styles.featuresSection}>
        <h2 className={styles.featuresSectionTitle}>Everything you need</h2>
        <div className={styles.featuresGrid}>
          {features.map((props, idx) => (
            <Feature key={idx} {...props} />
          ))}
        </div>
      </section>

      <CookieConsent>This website uses cookies to enhance the user experience.</CookieConsent>
    </Layout>
  );
}
