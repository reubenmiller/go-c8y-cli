import React from 'react';
import classnames from 'classnames';
import Layout from '@theme/Layout';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import useBaseUrl from '@docusaurus/useBaseUrl';
import CookieConsent from "react-cookie-consent";

import styles from './styles.module.css';

const features = [
  {
    title: 'Automate',
    imageUrl: 'img/undraw_docusaurus_mountain.svg',
    description: (
      <>
        Command line tools to automated common tasks
      </>
    ),
  },
  {
    title: 'Native pipeline support',
    imageUrl: 'img/undraw_docusaurus_tree.svg',
    description: (
      <>
        Chain multiple commands together and pass data to other 3rd party tools
      </>
    ),
  },
  {
    title: 'activity log',
    imageUrl: 'img/undraw_docusaurus_react.svg',
    description: (
      <>
        Track interactions with Cumulocity for improved traceability
      </>
    ),
  },
  {
    title: 'activity log',
    imageUrl: 'img/undraw_docusaurus_react.svg',
    description: (
      <>
        Track interactions with Cumulocity for improved traceability
      </>
    ),
  },
  {
    title: 'activity log',
    imageUrl: 'img/undraw_docusaurus_react.svg',
    description: (
      <>
        Track interactions with Cumulocity for improved traceability
      </>
    ),
  },
];

function Feature({imageUrl, title, description}) {
  const imgUrl = useBaseUrl(imageUrl);
  return (
    <div className={classnames('col col--4', styles.feature)}>
      {imgUrl && (
        <div className="text--center">
          <img className={styles.featureImage} src={imgUrl} alt={title} />
        </div>
      )}
      <h3>{title}</h3>
      <p>{description}</p>
    </div>
  );
}

export default function Home() {
  const context = useDocusaurusContext();
  const {siteConfig = {}} = context;

  return (
    <Layout
      title={`${siteConfig.title}`}
      description="go-c8y-cli documentation">
      <div className={styles.hero}>
        <header>
          <h1>{siteConfig.title}</h1>
          <p>{siteConfig.tagline}</p>
          <div className={styles.buttons}>
            <Link to={useBaseUrl('docs/')}>Get Started</Link>
          </div>
          <div className={styles.buttons}>
            <Link to={useBaseUrl('docs/cli/')}>API Documentation</Link>
          </div>
          <CookieConsent>This website uses cookies to enhance the user experience.</CookieConsent>
        </header>
        <main>
          {features && features.length > 0 && (
            <section className={styles.section}>
              <div className={styles.features}>
                {features.map((props, idx) => (
                  <Feature key={idx} {...props} />
                ))}
              </div>
            </section>
          )}
        </main>
      </div>
    </Layout>
  );
}
