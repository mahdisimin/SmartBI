-- Registers the Synops dashboard as an internal app route (not an external
-- URL, per the convention: Link starting with "/" = internal SPA route,
-- rendered without opening a new tab; anything else = external link).

IF NOT EXISTS (SELECT 1 FROM APP.WebAppLinkList WHERE Link = '/dashboards/synops')
BEGIN
    INSERT INTO APP.WebAppLinkList (Name, Link)
    VALUES ('Synops Activity', '/dashboards/synops');
END
GO

-- Grants every existing user access to it (idempotent — safe to re-run).
-- To grant only specific users instead, replace the SELECT below with a
-- literal list, e.g.:
--   INSERT INTO APP.User_WebAppLink (UserID, LinkID)
--   SELECT v.UserID, l.ID
--   FROM (VALUES (1), (2)) AS v(UserID)
--   CROSS JOIN APP.WebAppLinkList l
--   WHERE l.Link = '/dashboards/synops';

INSERT INTO APP.User_WebAppLink (UserID, LinkID)
SELECT u.ID, l.ID
FROM APP.[USER] u
CROSS JOIN APP.WebAppLinkList l
WHERE l.Link = '/dashboards/synops'
  AND NOT EXISTS (
      SELECT 1 FROM APP.User_WebAppLink existing
      WHERE existing.UserID = u.ID AND existing.LinkID = l.ID
  );
GO
