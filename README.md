<!-- Improved compatibility of back to top link: See: https://github.com/othneildrew/Best-README-Template/pull/73 -->
<a name="readme-top"></a>

<!-- PROJECT SHIELDS 
[![Contributors][contributors-shield]][contributors-url]
[![Forks][forks-shield]][forks-url]
[![Stargazers][stars-shield]][stars-url]
[![Issues][issues-shield]][issues-url]
[![LinkedIn][linkedin-shield]][linkedin-url] -->



<!-- PROJECT LOGO -->
<br />
<div align="center">
  <a href="https://movinglake.com/">
    <img src="https://www.movinglake.com/docs/images/movinglakelogo-img.svg" alt="Logo" width="80" height="80">
  </a>

  <h3 align="center">Postgres Webhook</h3>

  <p align="center">
    Run webhooks directly from your Postgres database using replication.
    <br />
    <a href="https://github.com/movinglake/pg_webhook"><strong>Explore the docs »</strong></a>
    <br />
    <br />
    <a href="https://www.loom.com/share/d3459bd391094aed9efa5c05912bd880">View Demo</a>
  </p>
</div>



<!-- TABLE OF CONTENTS -->
<details>
  <summary>Table of Contents</summary>
  <ol>
    <li>
      <a href="#about-the-project">About The Project</a>
      <ul>
        <li><a href="#built-with">Built With</a></li>
      </ul>
    </li>
    <li>
      <a href="#postgres-setup">Postgres Setup</a></li>
      <ul>
        <li><a href="#standalon-postgres">Docker</a></li>
        <li><a href="#rds-postgres">Compile and run</a></li>
      </ul>
    <li>
      <a href="#getting-started">Getting Started</a>
      <ul>
        <li><a href="#docker">Docker</a></li>
        <li><a href="#compile-and-run">Compile and run</a></li>
      </ul>
    </li>
    <li><a href="#system">System</a></li>
    <li><a href="#license">License</a></li>
    <li><a href="#contact">Contact</a></li>
    <li><a href="#managed-version">Managed Version</a></li>
    <li><a href="#acknowledgments">Acknowledgments</a></li>
  </ol>
</details>



<!-- ABOUT THE PROJECT -->
## About The Project

This project came about from MovingLake technological development. Webhooks directly from a database using a replication slot was one of our core desires since before we founded MovingLake. This is now a reality and we wanted to give back to the awesome Open-Source Community by creating this Repo. Hope you enjoy it!

<p align="right">(<a href="#readme-top">back to top</a>)</p>



### Built With

* [![Golang][Go.dev]][Go-url]
* [![Sqlite][Sqlite.db]][Sqlite-url]

<p align="right">(<a href="#readme-top">back to top</a>)</p>

## Postgres Setup

Pg_webhook uses a logical replication slot to get real-time data from the main database. As such a few configuration steps must be followed to allow this to happen

### Standalone Postgres

Use the following commands to get started. We'll create a database, a user and setup postgres to create a logical replication slot.

```
create database pg_webhook;
create user pg_webhook with replication password 'pg_webhook';
```

Next you need to add some lines to Postgres' configuration files. If you do not know where this files are, you can run `ps -ef | grep postgres`. You shoud see a line such as:

```
/opt/homebrew/opt/postgresql/bin/postgres -D /opt/homebrew/var/postgres
```

The last piece of this line (`/opt/homebrew/var/postgres`)  is where your postgres configuration files will be stored. Now open `pg_hba.conf` and add the following line:

```
host replication pg_webhook 127.0.0.1/32 md5
```

Next open `postgres.conf` and add these following lines:

```
wal_level=logical
max_wal_senders=5
max_replication_slots=5
```

Finally when specifying the postgres DNS string when running pg_webhook, make sure it has the replication query parameter `?replication=database`. Eg `postgres://pg_webhook:pg_webhook@localhost:5432/pg_webhook?replication=database`


### RDS Postgres

RDS does not let you run with a real superuser, and also doesn't let you change the configuration files. Most likely because of multitenant systems. To circumvent this, the easiest way to go is to use Postgres extension `pglogical`.

Please follow [this link](https://aws.amazon.com/blogs/database/part-2-upgrade-your-amazon-rds-for-postgresql-database-using-the-pglogical-extension/) to set-up your RDS instance with the extension. Also checkout [this other link](https://aws.amazon.com/premiumsupport/knowledge-center/rds-postgresql-resolve-preload-error/) if don't know how to turn on the preloaded libraries for RDS.

Once you follow these steps, try and follow the rest from #standalone-postgres

<!-- GETTING STARTED -->
## Getting Started


### Docker

The easiest way to get started is to build the Dockerfile 

`docker build -t latest .`

and then run the image with the appropriate environment variables

`docker run --rm -it --net=host -e PG_DNS='postgres://postgres@localhost:5432/postgres?replication=database' -e WEBHOOK_SERVICE_URL='http://localhost:9000/' latest`

### Compile and run

You can also run directly the extractor, just use

`go run main.go`

Do ensure that the environment variables are set beforehand.

## System

This extractor runs on Golang. It uses the pglogrepl library to create the replication slots and subscription. It then uses a producer consumer pattern to send received messages to the downstream service. If the downstream service cannot receive the requests, this code retries automatically for upto a day. After that it logs the failed requests into a Sqlite database on disk.

Consumer concurrency can be set higher if a high volume of transactions are being sent to the database.

<!-- LICENSE -->
## License

Distributed under the GNU AGPL 3 License. See `LICENSE` for more information.

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- CONTACT -->
## Contact

MovingLake - [@MovingLake](https://twitter.com/MovingLake) - info@movinglake.com

Project Link: [https://github.com/movinglake/pg_webhook](https://github.com/movinglake/pg_webhook)

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- MANAGED VERSION -->
## Managed Version

If you ever want to have a managed version of this system built for scale and completely maintenance free then we'd be happy to accommodate your requests. Please use this [link](app.movinglake.com) to create a Postgres connector with a webhook destination.

<p align="right">(<a href="#readme-top">back to top</a>)</p>


## Install in Kubernetes via Helm

Simply add our helm chart repository:

```
helm repo add movinglake https://movinglake.github.io/helmcharts/
```

Then install the chart:

```
helm upgrade --install my-pg-webhook \
  --namespace my-namespace \
  --set postgres.url="postgres://user:password@hostname:5432/postgres?replication=database" \
  --set webhook.url="https://my-webhook.com" \
  charts/pg-webhook
```

<!-- ACKNOWLEDGMENTS -->
## Acknowledgments

Want to acknowledge the pglogrepl repo and its creators for such a great tool!

* [pglogrepl](https://github.com/jackc/pglogrepl)
<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- MARKDOWN LINKS & IMAGES -->
<!-- https://www.markdownguide.org/basic-syntax/#reference-style-links -->
[contributors-shield]: https://img.shields.io/github/contributors/othneildrew/Best-README-Template.svg?style=for-the-badge
[contributors-url]: https://github.com/othneildrew/Best-README-Template/graphs/contributors
[forks-shield]: https://img.shields.io/github/forks/othneildrew/Best-README-Template.svg?style=for-the-badge
[forks-url]: https://github.com/othneildrew/Best-README-Template/network/members
[stars-shield]: https://img.shields.io/github/stars/othneildrew/Best-README-Template.svg?style=for-the-badge
[stars-url]: https://github.com/othneildrew/Best-README-Template/stargazers
[issues-shield]: https://img.shields.io/github/issues/othneildrew/Best-README-Template.svg?style=for-the-badge
[issues-url]: https://github.com/othneildrew/Best-README-Template/issues
[linkedin-shield]: https://img.shields.io/badge/-LinkedIn-black.svg?style=for-the-badge&logo=linkedin&colorB=555
[linkedin-url]: https://linkedin.com/in/movinglake
[Go-url]: https://go.dev/
[Go.dev]: https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white
[Sqlite-url]: https://www.sqlite.org/index.html
[Sqlite.db]: https://img.shields.io/badge/sqlite-%2307405e.svg?style=for-the-badge&logo=sqlite&logoColor=white