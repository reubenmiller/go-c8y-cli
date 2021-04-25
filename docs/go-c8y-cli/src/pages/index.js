import React from 'react';
import classnames from 'classnames';
import Layout from '@theme/Layout';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import useBaseUrl from '@docusaurus/useBaseUrl';
import CookieConsent from "react-cookie-consent";
import Video from '@site/src/components/video';

import styles from './styles.module.css';

const features = [
  {
    title: 'Automate',
    imageUrl: 'img/automate.svg',
    description: (
      <>
        Easy to write scripts to automate tasks
      </>
    ),
  },
  {
    title: 'Native pipeline support',
    imageUrl: 'img/pipeline.svg',
    description: (
      <>
        Chain multiple commands together and pass data to other 3rd party tools
      </>
    ),
  },
  {
    title: 'Activity log',
    imageUrl: 'img/activitylog.svg',
    description: (
      <>
        Track interactions with Cumulocity for improved traceability
      </>
    ),
  },
  {
    title: 'Workers',
    imageUrl: 'img/workers.svg',
    description: (
      <>
        Controlled concurrency to send multiple requests at the same time
      </>
    ),
  },
  {
    title: 'Highly Configurable',
    imageUrl: 'img/config.svg',
    description: (
      <>
        Customize to your needs
      </>
    ),
  },
  // {
  //   title: 'activity log',
  //   imageUrl: 'img/undraw_docusaurus_react.svg',
  //   description: (
  //     <>
  //       Track interactions with Cumulocity for improved traceability
  //     </>
  //   ),
  // },
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

const mainStyle = {
  width: '100%',
};

export default function Home() {
  const context = useDocusaurusContext();
  const {siteConfig = {}} = context;

  return (
    <Layout
      title={`${siteConfig.title}`}
      description="go-c8y-cli documentation">
      <div className={styles.hero}>
        <header className={styles.pageHeader}>
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
        <main style={mainStyle}>
          <section>
            <Video
              videoSrcURL="https://asciinema.org/a/326455/embed?speed=1.2&autoplay=false&size=small&rows=30"
              videoTitle="PSc8y PowerShell demonstration"
              width="90%"
              height="600px"
              ></Video>
          </section>
        
          {features && features.length > 0 && (
            <section className={styles.section}>
              <div className={styles.features}>
                {features.map((props, idx) => (
                  <Feature key={idx} className={styles.feature} {...props} />
                ))}
              </div>
            </section>
          )}
        </main>
      </div>
    </Layout>
  );
}
