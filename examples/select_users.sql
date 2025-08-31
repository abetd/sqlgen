SELECT
  *
FROM
  users
WHERE
  id = /** int .ID **/1234
  AND (
    name = /** string .Name **/'John Doe'
  OR
    name IN /** in .Names **/('Foo Bar', 'Bar Baz')
  )
