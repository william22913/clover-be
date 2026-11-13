-- FEAT-001 v1 catalogue: 13 films in the approved mock's order.
-- Idempotent by slug: a second start writes nothing. The conflict target
-- repeats the partial index predicate (WHERE NOT is_deleted) or inference fails.

INSERT INTO movies (slug, title, short_description, release_year, runtime_seconds, poster_file, still_file, video_file, display_order)
VALUES
 ('halcyon', 'Halcyon Bay', 'A marine biologist comes back to the town that took her brother, and finds the tide still keeping his secrets.', 2023, 7560, 'p-halcyon.jpg', 'still-halcyon.jpg', 'clover-sample.mp4', 1),
 ('quiet', 'The Quiet Province', 'A schoolteacher keeps a village’s last winter alive by writing letters no one will ever post.', 2025, 6240, 'p-quiet.jpg', 'still-quiet.jpg', 'clover-sample.mp4', 2),
 ('vesper', 'Vesper Light', 'Two strangers miss the same last ferry and spend one night deciding whether to miss it on purpose.', 2022, 6660, 'p-vesper.jpg', 'still-vesper.jpg', 'clover-sample.mp4', 3),
 ('cartographer', 'The Cartographer', 'One woman sets out to map the last unmapped valley in Europe, one ridge at a time.', 2024, 7080, 'p-cartographer.jpg', 'still-quiet.jpg', 'clover-sample.mp4', 4),
 ('meridian', 'Meridian Line', 'A night train, a stolen passport, and eight hours to the border.', 2021, 7920, 'p-meridian.jpg', 'still-vesper.jpg', 'clover-sample.mp4', 5),
 ('glasshouse', 'Glasshouse Summer', 'Three siblings, one greenhouse, and the summer their mother’s will is read.', 2024, 5940, 'p-glasshouse.jpg', 'still-halcyon.jpg', 'clover-sample.mp4', 6),
 ('fathoms', 'Nine Fathoms', 'A deep-sea station loses contact with the surface for eleven minutes.', 2023, 6420, 'p-fathoms.jpg', 'still-halcyon.jpg', 'clover-sample.mp4', 7),
 ('lowcountry', 'Low Country', 'A family of shrimpers fights the water, the bank, and each other for one season.', 2022, 7440, 'p-lowcountry.jpg', 'still-quiet.jpg', 'clover-sample.mp4', 8),
 ('saltyear', 'The Salt Year', 'A year on the road with a band that never made it and never stopped.', 2025, 5760, 'p-saltyear.jpg', 'still-vesper.jpg', 'clover-sample.mp4', 9),
 ('lanterns', 'Paper Lanterns', 'A projectionist falls for the woman who returns to the same empty seat.', 2024, 6120, 'p-lanterns.jpg', 'still-vesper.jpg', 'clover-sample.mp4', 10),
 ('harbour', 'Harbour Lights', 'The last show at a seaside theatre, on the night the lights went out.', 2021, 6900, 'p-harbour.jpg', 'still-halcyon.jpg', 'clover-sample.mp4', 11),
 ('understudy', 'The Understudy', 'The worst actor in the company gets one night on stage. He has not learned the lines.', 2023, 5460, 'p-understudy.jpg', 'still-quiet.jpg', 'clover-sample.mp4', 12),
 ('static', 'Static', 'The old Scala shows something different every night, and nobody runs the projector.', 2025, 6540, 'p-static.jpg', 'still-vesper.jpg', 'clover-sample.mp4', 13)
ON CONFLICT (slug) WHERE NOT is_deleted DO UPDATE SET
    title = excluded.title,
    short_description = excluded.short_description,
    release_year = excluded.release_year,
    runtime_seconds = excluded.runtime_seconds,
    poster_file = excluded.poster_file,
    still_file = excluded.still_file,
    video_file = excluded.video_file,
    display_order = excluded.display_order,
    updated_at = now()
WHERE (movies.title, movies.short_description, movies.release_year, movies.runtime_seconds,
       movies.poster_file, movies.still_file, movies.video_file, movies.display_order)
    IS DISTINCT FROM (excluded.title, excluded.short_description, excluded.release_year, excluded.runtime_seconds,
       excluded.poster_file, excluded.still_file, excluded.video_file, excluded.display_order);
