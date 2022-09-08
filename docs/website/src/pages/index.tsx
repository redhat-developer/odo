import React from 'react';
import clsx from 'clsx';
import Layout from '@theme/Layout';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import styles from './index.module.css';
import HomepageFeatures from '../components/HomepageFeatures';

import useBaseUrl from '@docusaurus/useBaseUrl';

export default function Home(): JSX.Element {
  const {
    siteConfig: {customFields, tagline},
  } = useDocusaurusContext();
  const {description} = customFields as {description: string};
  const {siteConfig} = useDocusaurusContext();
  return (
    <Layout title={tagline} description={description}>
      <main>
        <div className={styles.banner}>
          <div className={styles.bannerInner}>
            <h1 className={styles.bannerProjectTagline}>
  <img
  alt='Logo'
  className={styles.logo}
  src={useBaseUrl('/img/logo.png')}
  />
              <span className={styles.bannerTitleTextHtml}>Fast <b>iterative</b> application <b>development</b> deployed immediately to your <b>kubernetes</b> cluster</span>
            </h1>
            <div className={styles.indexCtas}>
              <Link className="button button--primary" to="/docs/introduction">
    Read the docs
              </Link>
              <Link className="button button--info" to="/docs/user-guides/quickstart/">
   Quickstart Guide 
              </Link>
              <span className={styles.indexCtasGitHubButtonWrapper}>
                <iframe
                  className={styles.indexCtasGitHubButton}
                  src="https://ghbtns.com/github-btn.html?user=redhat-developer&amp;repo=odo&amp;type=star&amp;count=true&amp;size=large"
                  width={160}
                  height={30}
                  title="GitHub Stars"
                  frameBorder={0}
                />
              </span>
            </div>
          </div>
        </div>
        <div className={clsx(styles.title, styles.titleDark)}>
          <div className={styles.titleInner}>
            Version 3 of odo is arriving 🚀<br></br> <Link to="/docs/overview/installation">Install</Link> and <Link to="/docs/user-guides/quickstart/">try out</Link> our new features ⭐️
          </div>
        </div>
        <div className={styles.overview}>
          <div className="container text--center margin-top--md">
            <div className="row">
              <div className="col col--5 col--offset-1">
                    <video className={styles.loopVideo} autoPlay loop muted><source src="/video/container_ship.webm" type="video/webm"/></video>
                <h2 className={clsx(styles.featureHeading)}>
                  Deploy on <b className={styles.kubernetesFont}>Kubernetes</b> and <b className={styles.openshiftFont}>OpenShift</b>
                </h2>
                <p className="padding-horiz--md">
    We provide first class support for both Kubernetes and OpenShift. Choose your favourite container orchestrator and deploy your application.
                </p>
              </div>
              <div className="col col--5">
                    <video className={styles.loopVideo} autoPlay loop muted><source src="/video/coding.webm" type="video/webm"/></video>
                <h2 className={clsx(styles.featureHeading)}>
                  Push code fast and often
                </h2>
                <p className="padding-horiz--md">
    Spend less time maintaining your deployment infrastructure and more time coding. Immediately have your application running each time you compile.
                </p>
              </div>
            </div>
          </div>
        </div>
        <div className={clsx(styles.overview, styles.overviewAlt)}>
          <div className="container text--center margin-top--lg">
            <div className="row">
              <div className="col">
                <img className={styles.featureImage} alt="foobar" src={useBaseUrl('/img/icons/client.png')}/>
                <h2 className={clsx(styles.featureHeading)}>
    Standalone client
                </h2>
                <p className="padding-horiz--md">
    <code>odo</code> is a standalone tool that communicates directly with the Kubernetes API. There is no requirement for a daemon or server process.
                </p>
              </div>
              <div className="col">
                <img className={styles.featureImage} alt="foobar" src={useBaseUrl('/img/icons/engineers.png')}/>
                <h2 className={clsx(styles.featureHeading)}>
                  Built for container engineers
                </h2>
                <p className="padding-horiz--md">
      Built from the ground up with application development on Kubernetes in mind. Each command has been carefuly crafted for application container development.
                </p>
              </div>
              <div className="col">
                <img className={styles.featureImage} alt="foobar" src={useBaseUrl('/img/icons/configuration.png')}/>
                <h2 className={clsx(styles.featureHeading)}>
                  No needed configuration
                </h2>
                <p className="padding-horiz--md">
    There is no need to dive into complex Kubernetes yaml configuration files. <code>odo</code> abstracts those concepts away and lets you focus on what matters most: code.
                </p>
              </div>
            </div>
          </div>
        </div>
        <div className={clsx(styles.overview, styles.overviewAlt)}>
          <div className="container text--center margin-bottom--md">
            <div className="row">
              <div className="col col--4 col--offset-2">
                <img className={styles.featureImage} alt="foobar" src={useBaseUrl('/img/icons/complex.png')}/>
                <h2 className={clsx(styles.featureHeading)}>
    Deploy a simple or complex application
                </h2>
                <p className="padding-horiz--md">
    Big or small, <code>odo</code> can deploy them all. Deploy a simple Node.JS application, or even a complex <Link to="https://github.com/operator-framework/">Operator</Link> backed service.
                </p>
              </div>
              <div className="col col--4">
                <img className={styles.featureImage} alt="foobar" src={useBaseUrl('/img/icons/tests.png')}/>
                <h2 className={clsx(styles.featureHeading)}>
                  Run your tests directly on the cluster
                </h2>
                <p className="padding-horiz--md">
    Debug and test remote applications deployed using <code>odo</code> directly from your IDE to Kubernetes. No more having to exit your IDE to push your application.
                </p>
              </div>
            </div>
          </div>
        </div>
      </main>
    </Layout>
  );
}
