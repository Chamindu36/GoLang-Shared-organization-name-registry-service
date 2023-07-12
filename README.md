# organization-name-registry-service

Organization name registry service for Choreo and other cloud services

### Technologies
GoLang, Azure, WSO2 Choreo, MySQL and CosomsDB

## Objective

Need to implement a service to share organization names across Ballerina Central, Choreo, Asgardeo and any other future WSO2 cloud products. The organization implementation can be different product to product but the org name should be unique across all the products.
This is similar to having a central domain/org name registry within WSO2. When a user tries to create an organization in a cloud product, the product can internally call the org-name registry service to check the availability first and then reserve the name upon org creation. The only concern occurs when ‘Yasith’ creates an org named ‘wso2’ in BC and then tries to create the same in Choreo. He needs a way to claim the org name and ideally this process should not delay the signup process.
One way to achieve this is keeping the org name against the user email/s in the org-name registry and allow users to reuse an org-name iff emails are the same.


## Scope description

Email will be used as the common attribute for identifying the owner of a org domain. Organization ownership claiming process will not be implemented in this.
