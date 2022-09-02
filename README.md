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
      <a href="#getting-started">Getting Started</a>
      <ul>
        <li><a href="#docker">Docker</a></li>
        <li><a href="#compile-and-run">Compile and run</a></li>
      </ul>
    </li>
    <li><a href="#system">System</a></li>
    <li><a href="#license">License</a></li>
    <li><a href="#contact">Contact</a></li>
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