SELECT
  *
FROM
  items
WHERE
  /** if false -**/
  -- hogehoge
  /**- end -**/
  1
  AND id = /** int .ID **/1234
  /**- if .IsSelectMultiNames **/
  AND ( /** multi .Where .Sep .Names -**/
    /**- if false -**/
    -- multi は
    -- Where = (name LIKE ? OR kana LIKE ?) を
    -- Names = {%foo%, %bar%) 分
    -- Sep = AND でつなげる
    (name LIKE '%foo%' OR kana LIKE '%foo%')
    AND
    (name LIKE '%bar%' OR kana LIKE '%bar%')
    /**- end -**/
  )
  AND ( /** multi "(name LIKE ? OR kana LIKE ?)" "AND" .Names -**/
    /**- if false -**/
    -- multi は クエリー部と接続部を文字列で指定できる
    (name LIKE '%foo%' OR kana LIKE '%foo%')
    AND
    (name LIKE '%bar%' OR kana LIKE '%bar%')
    /**- end -**/
  )
  /**- end **/
