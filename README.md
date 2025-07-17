# MDS - AWIPS

Services for ingesting, parsing, and storing AWIPS products to the MDS.

## Background

The National Weather Service has long been using AWIPS to produce and disseminate data between offices, emergency management, and the public. AWIPS products are sent as text data across several transmission methods.
One of these transmission methods is the NOAA Weather Wire Service, an XMPP system that provides users with access to real-time publications of NOAA text products. Hence, using the NWWS as a primary data source,
MDS AWIPS ingests text data from the NWWS and parses the products into useful data.

The main reason this repo uses the name AWIPS, instead of something like NWSText, is because we honestly only care to store products that were produced by AWIPS. These products include but are not limited to:

- Warnings, Watches, Advisoring (WWA)
- SPC products (MCD, outlooks, storm reports)
- Special Weather Statements
- NHC Hurricane updates/outlooks
- NWS/NOAA administrative messages/updates

AWIPS was also preceeded by AFOS, another system for producing and disseminating products. However, the IEM seems to have nailed archiving this so we will leave the glory to them. The focus for us can stay on AWIPS.

## Design

The design of this module or package or system (whatever you want to call it) is still up for debate. Sitting in a room by yourself contemplating the design of such a large system leaves a lot of the burden of problem
solving to a single very inexperienced developer.

The current thinking is that this is a module of the MDS, since the MDS database is a completely separate repository. However, some thought has gone into whether a separate database system should be spooled up entirely
for the purposes of archiving this data while allowing other services to access this data through APIs. The latter might honestly be the way to go but we will have to see.

## Acknowledgements

Firstly, huge shoutout to Daryl and the IEM for being the main inspiration for this. Without his code and wealth of knowledge this project would not exist. We whole heartedly respect the IEM and all the work that has been put into it.
We do not wish for this project to replace the IEM but merely collaborate with it. That is why if you need high quality archived data, go to the IEM first and here second.
In the long run, it would be nice to be able to sync the IEM and MDS for more accurate archiving.

Next, we have to acknowledge NOAA and the NWS for allowing their products to be freely available in real-time. It is so hard to find any other national weather authority that has such an abundance of open data. So to the
US and NOAA we are thankful.
