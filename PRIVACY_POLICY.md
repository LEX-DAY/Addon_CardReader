# Privacy Policy for Card Reader Bridge

Effective date: February 24, 2026

## 1. Overview
Card Reader Bridge is a browser extension that receives card reader data from a locally installed native host and inserts the scanned card value into the currently focused input field on the active web page.

## 2. Data We Collect
We do not collect personal data from users.

The extension does not collect or store:
- names, emails, phone numbers, addresses, or account identifiers;
- passwords, PINs, or authentication secrets;
- browsing history for analytics purposes;
- page content for transmission to external servers.

## 3. How Data Is Processed
- Card scan data is processed locally on the user device.
- The extension uses Chrome Native Messaging to communicate with a local native application.
- The scanned value is inserted into the focused form field on the active tab.

## 4. Data Sharing
We do not sell, transfer, or share user data with third parties.

No card scan data is sent to our servers.

## 5. Data Retention
We do not retain user data on remote systems.

Any transient processing occurs locally in memory only for the purpose of immediate input insertion.

## 6. Permissions Usage
- `nativeMessaging`: required to receive data from the local card reader host.
- `tabs` and `activeTab`: required to target the active tab and insert the value into the focused field.
- Host access (content scripts): required to work on user web pages where input is needed.

## 7. Remote Code
We do not use remote code.

All extension code is packaged with the extension. No external scripts, remote modules, or remotely hosted Wasm are executed.

## 8. Children's Privacy
This extension is not directed to children under 13, and we do not knowingly collect children's personal data.

## 9. Changes to This Policy
We may update this Privacy Policy from time to time. Any update will be reflected by changing the effective date in this document.

## 10. Contact
For privacy questions, contact the publisher through the support contact listed on the extension page.
