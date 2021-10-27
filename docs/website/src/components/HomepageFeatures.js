import React from 'react';
import clsx from 'clsx';
import styles from './HomepageFeatures.module.css';

const FeatureList = [
  {
    title: 'Daemonless',
    Svg: 'daemonless',
    description: (
      <>
        <code>odo</code> is a standalone CLI tool; it communicates directly with the Kubernetes API.
          <code>odo</code> doesn't require you to run a daemon (server) process.
          All you need is to download the binary and start playing with it.
      </>
    ),
  },
  {
    title: 'No more YAML!',
    Svg: 'noyaml',
    description: (
      <>
        <code>odo</code> doesn't <i>require</i> you to work with YAML file.
          It uses constructs of its own and helps developers focus on their application development instead of having to learn Kubernetes.
          But if you do want to play with the YAML, you absolutely can! <code>odo</code> uses devfile format and you can always modify a devfile manually if you would really like to.
      </>
    ),
  },
  {
    title: 'Made for Application Developers',
    Svg: 'appdevs',
    description: (
      <>
        <code>odo</code> is made for Application Developers and Architects.
          It abstracts Kubernetes constructs into its own constructs which are more aligned with application development than Ops.
          <code>kubectl</code> is more focussed on Ops; <code>odo</code> is <i>totally</i> focussed on Devs.
      </>
    ),
  },
];

function Feature({Svg, title, description}) {
  return (
    <div className={clsx('col col--4')}>
      <div className="text--left">
        <Svg className={styles.featureSvg} alt={title} />
      </div>
      <div className="text--left padding-horiz--md">
        <h3>{title}</h3>
        <p>{description}</p>
      </div>
    </div>
  );
}

export default function HomepageFeatures() {
  return (
    <section className={styles.features}>
      <div className="container">
        <div className="row">
          {FeatureList.map((props, idx) => (
            <Feature key={idx} {...props} />
          ))}
        </div>
      </div>
    </section>
  );
}
